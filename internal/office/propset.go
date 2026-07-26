package office

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
	"unicode/utf16"
)

// Property identifiers of the SummaryInformation property set.
const (
	pidCodePage   = 0x00000001
	pidAuthor     = 0x00000004
	pidLastAuthor = 0x00000008
)

const (
	vtI2     = 2
	vtLPSTR  = 30
	vtLPWSTR = 31
)

var errBadPropertySet = errors.New("повреждённый блок свойств документа")

type property struct {
	id  uint32
	raw []byte
}

type propertySet struct {
	header     []byte // everything before the first section
	sectionOff int
	codePage   uint16
	props      []property
}

func parsePropertySet(data []byte) (*propertySet, error) {
	if len(data) < 48 || binary.LittleEndian.Uint16(data[:2]) != 0xFFFE {
		return nil, errBadPropertySet
	}
	if binary.LittleEndian.Uint32(data[24:28]) != 1 {
		// Rewriting would invalidate the offsets of any extra sections.
		return nil, fmt.Errorf("%w: несколько секций свойств", errBadPropertySet)
	}
	sectionOff := int(binary.LittleEndian.Uint32(data[44:48]))
	if sectionOff+8 > len(data) {
		return nil, errBadPropertySet
	}
	section := data[sectionOff:]
	size := int(binary.LittleEndian.Uint32(section[0:4]))
	count := int(binary.LittleEndian.Uint32(section[4:8]))
	if size > len(section) || count < 0 || 8+count*8 > len(section) {
		return nil, errBadPropertySet
	}

	type entry struct {
		id     uint32
		offset int
	}
	entries := make([]entry, 0, count)
	for i := 0; i < count; i++ {
		base := 8 + i*8
		entries = append(entries, entry{
			id:     binary.LittleEndian.Uint32(section[base : base+4]),
			offset: int(binary.LittleEndian.Uint32(section[base+4 : base+8])),
		})
	}
	ordered := append([]entry(nil), entries...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].offset < ordered[j].offset })

	bounds := make(map[uint32]int, len(ordered))
	for i, e := range ordered {
		end := size
		if i+1 < len(ordered) {
			end = ordered[i+1].offset
		}
		if e.offset < 0 || end > len(section) || end < e.offset {
			return nil, errBadPropertySet
		}
		bounds[e.id] = end
	}

	ps := &propertySet{header: append([]byte(nil), data[:sectionOff]...), sectionOff: sectionOff}
	for _, e := range entries {
		raw := append([]byte(nil), section[e.offset:bounds[e.id]]...)
		ps.props = append(ps.props, property{id: e.id, raw: raw})
		if e.id == pidCodePage && len(raw) >= 6 {
			ps.codePage = binary.LittleEndian.Uint16(raw[4:6])
		}
	}
	return ps, nil
}

func (ps *propertySet) set(id uint32, raw []byte) {
	for i := range ps.props {
		if ps.props[i].id == id {
			ps.props[i].raw = raw
			return
		}
	}
	ps.props = append(ps.props, property{id: id, raw: raw})
}

// setString stores a string property using the encoding of the property set.
func (ps *propertySet) setString(id uint32, value string) {
	if ps.codePage == 1200 {
		ps.set(id, encodeLPWSTR(value))
		return
	}
	if !isASCII(value) {
		ps.codePage = 65001
		ps.setCodePage(65001)
	}
	ps.set(id, encodeLPSTR(value))
}

func (ps *propertySet) setCodePage(cp uint16) {
	raw := make([]byte, 8)
	binary.LittleEndian.PutUint32(raw[0:4], vtI2)
	binary.LittleEndian.PutUint16(raw[4:6], cp)
	ps.set(pidCodePage, raw)
}

func encodeLPSTR(value string) []byte {
	body := append([]byte(value), 0)
	raw := make([]byte, 8, 8+len(body)+3)
	binary.LittleEndian.PutUint32(raw[0:4], vtLPSTR)
	binary.LittleEndian.PutUint32(raw[4:8], uint32(len(body)))
	raw = append(raw, body...)
	for len(raw)%4 != 0 {
		raw = append(raw, 0)
	}
	return raw
}

func encodeLPWSTR(value string) []byte {
	units := append(utf16.Encode([]rune(value)), 0)
	raw := make([]byte, 8, 8+len(units)*2+2)
	binary.LittleEndian.PutUint32(raw[0:4], vtLPWSTR)
	binary.LittleEndian.PutUint32(raw[4:8], uint32(len(units)))
	for _, unit := range units {
		raw = binary.LittleEndian.AppendUint16(raw, unit)
	}
	for len(raw)%4 != 0 {
		raw = append(raw, 0)
	}
	return raw
}

func isASCII(value string) bool {
	for i := 0; i < len(value); i++ {
		if value[i] > 0x7F {
			return false
		}
	}
	return true
}

// bytes serialises the property set back into its binary representation.
func (ps *propertySet) bytes() []byte {
	// The code page property must come first so readers can decode strings.
	sort.SliceStable(ps.props, func(i, j int) bool {
		return ps.props[i].id == pidCodePage && ps.props[j].id != pidCodePage
	})

	count := len(ps.props)
	valuesStart := 8 + count*8
	table := make([]byte, 8+count*8)
	binary.LittleEndian.PutUint32(table[4:8], uint32(count))

	values := make([]byte, 0, 256)
	for i, prop := range ps.props {
		for len(values)%4 != 0 {
			values = append(values, 0)
		}
		binary.LittleEndian.PutUint32(table[8+i*8:12+i*8], prop.id)
		binary.LittleEndian.PutUint32(table[12+i*8:16+i*8], uint32(valuesStart+len(values)))
		values = append(values, prop.raw...)
	}
	section := append(table, values...)
	binary.LittleEndian.PutUint32(section[0:4], uint32(len(section)))

	out := append([]byte(nil), ps.header...)
	binary.LittleEndian.PutUint32(out[44:48], uint32(ps.sectionOff))
	if len(out) < ps.sectionOff {
		out = append(out, make([]byte, ps.sectionOff-len(out))...)
	}
	return append(out[:ps.sectionOff], section...)
}
