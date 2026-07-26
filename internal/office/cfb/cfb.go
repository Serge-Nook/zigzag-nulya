// Package cfb provides minimal read/write access to Compound File Binary
// containers (the format used by legacy .doc/.xls/.ppt documents).
package cfb

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"unicode/utf16"
)

const (
	freeSect    = 0xFFFFFFFF
	endOfChain  = 0xFFFFFFFE
	maxRegSect  = 0xFFFFFFFA
	dirEntrySz  = 128
	miniSectSz  = 64
	entryStream = 2
)

var signature = []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}

// ErrNotCFB indicates that the data is not a compound file.
var ErrNotCFB = errors.New("файл не является документом формата OLE (CFB)")

// ErrNoSpace indicates that a stream cannot grow inside the container.
var ErrNoSpace = errors.New("недостаточно свободного места в контейнере для новых данных")

// File is an in-memory compound file that can be modified and written back.
type File struct {
	data       []byte
	sectorSize int
	fat        []uint32
	miniFAT    []uint32
	dir        []dirEntry
	cutoff     uint32
}

type dirEntry struct {
	offset int // byte offset of the entry inside data
	Name   string
	Type   byte
	Start  uint32
	Size   uint64
}

// Open parses a compound file held in memory.
func Open(data []byte) (*File, error) {
	if len(data) < 512 || !bytes.Equal(data[:8], signature) {
		return nil, ErrNotCFB
	}
	sectorShift := binary.LittleEndian.Uint16(data[30:32])
	if sectorShift != 9 && sectorShift != 12 {
		return nil, fmt.Errorf("%w: неизвестный размер сектора", ErrNotCFB)
	}
	f := &File{
		data:       data,
		sectorSize: 1 << sectorShift,
		cutoff:     binary.LittleEndian.Uint32(data[56:60]),
	}
	if err := f.readFAT(); err != nil {
		return nil, err
	}
	if err := f.readMiniFAT(); err != nil {
		return nil, err
	}
	return f, f.readDirectory()
}

// Data returns the current contents of the compound file.
func (f *File) Data() []byte { return f.data }

func (f *File) sectorOffset(sector uint32) int {
	return (int(sector) + 1) * f.sectorSize
}

func (f *File) sector(sector uint32) ([]byte, error) {
	start := f.sectorOffset(sector)
	if start < 0 || start+f.sectorSize > len(f.data) {
		return nil, fmt.Errorf("%w: сектор %d вне файла", ErrNotCFB, sector)
	}
	return f.data[start : start+f.sectorSize], nil
}

func (f *File) readFAT() error {
	difat := make([]uint32, 0, 109)
	for i := 0; i < 109; i++ {
		difat = append(difat, binary.LittleEndian.Uint32(f.data[76+i*4:80+i*4]))
	}
	next := binary.LittleEndian.Uint32(f.data[68:72])
	for next <= maxRegSect {
		sec, err := f.sector(next)
		if err != nil {
			return err
		}
		count := f.sectorSize/4 - 1
		for i := 0; i < count; i++ {
			difat = append(difat, binary.LittleEndian.Uint32(sec[i*4:i*4+4]))
		}
		next = binary.LittleEndian.Uint32(sec[count*4:])
	}
	for _, fatSector := range difat {
		if fatSector > maxRegSect {
			continue
		}
		sec, err := f.sector(fatSector)
		if err != nil {
			return err
		}
		for i := 0; i < f.sectorSize/4; i++ {
			f.fat = append(f.fat, binary.LittleEndian.Uint32(sec[i*4:i*4+4]))
		}
	}
	return nil
}

func (f *File) readMiniFAT() error {
	next := binary.LittleEndian.Uint32(f.data[60:64])
	for next <= maxRegSect {
		sec, err := f.sector(next)
		if err != nil {
			return err
		}
		for i := 0; i < f.sectorSize/4; i++ {
			f.miniFAT = append(f.miniFAT, binary.LittleEndian.Uint32(sec[i*4:i*4+4]))
		}
		next = f.nextInFAT(next)
	}
	return nil
}

func (f *File) nextInFAT(sector uint32) uint32 {
	if int(sector) >= len(f.fat) {
		return endOfChain
	}
	return f.fat[sector]
}

func (f *File) readDirectory() error {
	sector := binary.LittleEndian.Uint32(f.data[48:52])
	for sector <= maxRegSect {
		base := f.sectorOffset(sector)
		if base+f.sectorSize > len(f.data) {
			return ErrNotCFB
		}
		for off := base; off+dirEntrySz <= base+f.sectorSize; off += dirEntrySz {
			raw := f.data[off : off+dirEntrySz]
			nameLen := int(binary.LittleEndian.Uint16(raw[64:66]))
			name := ""
			if nameLen >= 2 && nameLen <= 64 {
				units := make([]uint16, 0, nameLen/2-1)
				for i := 0; i < nameLen-2; i += 2 {
					units = append(units, binary.LittleEndian.Uint16(raw[i:i+2]))
				}
				name = string(utf16.Decode(units))
			}
			f.dir = append(f.dir, dirEntry{
				offset: off,
				Name:   name,
				Type:   raw[66],
				Start:  binary.LittleEndian.Uint32(raw[116:120]),
				Size:   binary.LittleEndian.Uint64(raw[120:128]),
			})
		}
		sector = f.nextInFAT(sector)
	}
	if len(f.dir) == 0 {
		return ErrNotCFB
	}
	return nil
}

func (f *File) find(name string) *dirEntry {
	for i := range f.dir {
		if f.dir[i].Type == entryStream && f.dir[i].Name == name {
			return &f.dir[i]
		}
	}
	return nil
}

// ReadStream returns the contents of the named stream.
func (f *File) ReadStream(name string) ([]byte, error) {
	entry := f.find(name)
	if entry == nil {
		return nil, fmt.Errorf("поток %q не найден", name)
	}
	if uint32(entry.Size) < f.cutoff {
		return f.readChain(entry.Start, int(entry.Size), true)
	}
	return f.readChain(entry.Start, int(entry.Size), false)
}

// HasStream reports whether the named stream exists.
func (f *File) HasStream(name string) bool { return f.find(name) != nil }

func (f *File) readChain(start uint32, size int, mini bool) ([]byte, error) {
	out := make([]byte, 0, size)
	sector := start
	for sector <= maxRegSect && len(out) < size {
		chunk, err := f.chunk(sector, mini)
		if err != nil {
			return nil, err
		}
		out = append(out, chunk...)
		if mini {
			sector = f.nextInMiniFAT(sector)
		} else {
			sector = f.nextInFAT(sector)
		}
	}
	if len(out) < size {
		return nil, fmt.Errorf("%w: поток короче заявленного размера", ErrNotCFB)
	}
	return out[:size], nil
}

func (f *File) nextInMiniFAT(sector uint32) uint32 {
	if int(sector) >= len(f.miniFAT) {
		return endOfChain
	}
	return f.miniFAT[sector]
}

// chunk returns the slice of the underlying data for one (mini) sector.
func (f *File) chunk(sector uint32, mini bool) ([]byte, error) {
	if !mini {
		return f.sector(sector)
	}
	root := f.dir[0]
	offset := int(sector) * miniSectSz
	current := root.Start
	for offset >= f.sectorSize && current <= maxRegSect {
		current = f.nextInFAT(current)
		offset -= f.sectorSize
	}
	sec, err := f.sector(current)
	if err != nil {
		return nil, err
	}
	if offset+miniSectSz > len(sec) {
		return nil, ErrNotCFB
	}
	return sec[offset : offset+miniSectSz], nil
}

// WriteStream replaces the contents of an existing stream. The stream may
// shrink; it may only grow while free (mini) sectors are available.
func (f *File) WriteStream(name string, content []byte) error {
	entry := f.find(name)
	if entry == nil {
		return fmt.Errorf("поток %q не найден", name)
	}
	mini := uint32(entry.Size) < f.cutoff && uint32(len(content)) < f.cutoff
	if !mini && uint32(entry.Size) < f.cutoff {
		return ErrNoSpace // migration mini -> regular is not supported
	}
	chunkSize := f.sectorSize
	if mini {
		chunkSize = miniSectSz
	}
	needed := (len(content) + chunkSize - 1) / chunkSize
	chain, err := f.chainSectors(entry.Start, mini)
	if err != nil {
		return err
	}
	for len(chain) < needed {
		next, err := f.allocate(mini)
		if err != nil {
			return err
		}
		if mini {
			if err := f.growMiniStream(next); err != nil {
				return err
			}
		}
		if len(chain) > 0 {
			f.setNext(chain[len(chain)-1], next, mini)
		} else {
			entry.Start = next
		}
		f.setNext(next, endOfChain, mini)
		chain = append(chain, next)
	}
	for i, sector := range chain {
		dst, err := f.chunk(sector, mini)
		if err != nil {
			return err
		}
		for j := range dst {
			dst[j] = 0
		}
		if i < needed {
			start := i * chunkSize
			end := start + chunkSize
			if end > len(content) {
				end = len(content)
			}
			copy(dst, content[start:end])
		}
	}
	if needed < len(chain) {
		f.setNext(chain[needed-1], endOfChain, mini)
		for _, sector := range chain[needed:] {
			f.setNext(sector, freeSect, mini)
		}
	}
	binary.LittleEndian.PutUint32(f.data[entry.offset+116:entry.offset+120], entry.Start)
	binary.LittleEndian.PutUint64(f.data[entry.offset+120:entry.offset+128], uint64(len(content)))
	entry.Size = uint64(len(content))
	return nil
}

func (f *File) chainSectors(start uint32, mini bool) ([]uint32, error) {
	var chain []uint32
	sector := start
	for sector <= maxRegSect {
		chain = append(chain, sector)
		if len(chain) > 1<<20 {
			return nil, fmt.Errorf("%w: зацикленная цепочка секторов", ErrNotCFB)
		}
		if mini {
			sector = f.nextInMiniFAT(sector)
		} else {
			sector = f.nextInFAT(sector)
		}
	}
	return chain, nil
}

// allocate reserves a free (mini) sector, extending the file if the free FAT
// slot points past the current end of data.
func (f *File) allocate(mini bool) (uint32, error) {
	table := f.fat
	if mini {
		table = f.miniFAT
	}
	for i, value := range table {
		if value != freeSect {
			continue
		}
		sector := uint32(i)
		if !mini {
			f.ensureSector(sector)
		}
		return sector, nil
	}
	return 0, ErrNoSpace
}

// ensureSector extends the underlying data so the sector exists.
func (f *File) ensureSector(sector uint32) {
	end := f.sectorOffset(sector) + f.sectorSize
	if end > len(f.data) {
		f.data = append(f.data, make([]byte, end-len(f.data))...)
	}
}

// growMiniStream makes sure the root mini stream covers the given mini sector.
func (f *File) growMiniStream(miniSector uint32) error {
	required := uint64(miniSector+1) * miniSectSz
	root := &f.dir[0]
	if required <= root.Size {
		return nil
	}
	chain, err := f.chainSectors(root.Start, false)
	if err != nil {
		return err
	}
	for uint64(len(chain)*f.sectorSize) < required {
		next, err := f.allocate(false)
		if err != nil {
			return err
		}
		sec, err := f.sector(next)
		if err != nil {
			return err
		}
		for i := range sec {
			sec[i] = 0
		}
		if len(chain) > 0 {
			f.setNext(chain[len(chain)-1], next, false)
		} else {
			root.Start = next
			binary.LittleEndian.PutUint32(f.data[root.offset+116:root.offset+120], next)
		}
		f.setNext(next, endOfChain, false)
		chain = append(chain, next)
	}
	root.Size = required
	binary.LittleEndian.PutUint64(f.data[root.offset+120:root.offset+128], required)
	return nil
}

func (f *File) setNext(sector, next uint32, mini bool) {
	if mini {
		if int(sector) < len(f.miniFAT) {
			f.miniFAT[sector] = next
			f.writeTableEntry(binary.LittleEndian.Uint32(f.data[60:64]), sector, next, true)
		}
		return
	}
	if int(sector) < len(f.fat) {
		f.fat[sector] = next
		f.writeFATEntry(sector, next)
	}
}

// writeTableEntry updates the on-disk MiniFAT entry for the given sector.
func (f *File) writeTableEntry(firstSector, sector, value uint32, mini bool) {
	if !mini {
		return
	}
	perSector := uint32(f.sectorSize / 4)
	current := firstSector
	index := sector
	for index >= perSector && current <= maxRegSect {
		current = f.nextInFAT(current)
		index -= perSector
	}
	sec, err := f.sector(current)
	if err != nil {
		return
	}
	binary.LittleEndian.PutUint32(sec[index*4:index*4+4], value)
}

// writeFATEntry updates the on-disk FAT entry for the given sector.
func (f *File) writeFATEntry(sector, value uint32) {
	perSector := uint32(f.sectorSize / 4)
	fatIndex := sector / perSector
	offset := sector % perSector
	fatSector, ok := f.fatSectorAt(fatIndex)
	if !ok {
		return
	}
	sec, err := f.sector(fatSector)
	if err != nil {
		return
	}
	binary.LittleEndian.PutUint32(sec[offset*4:offset*4+4], value)
}

func (f *File) fatSectorAt(index uint32) (uint32, bool) {
	if index < 109 {
		return binary.LittleEndian.Uint32(f.data[76+index*4 : 80+index*4]), true
	}
	perSector := uint32(f.sectorSize/4) - 1
	next := binary.LittleEndian.Uint32(f.data[68:72])
	index -= 109
	for next <= maxRegSect {
		sec, err := f.sector(next)
		if err != nil {
			return 0, false
		}
		if index < perSector {
			return binary.LittleEndian.Uint32(sec[index*4 : index*4+4]), true
		}
		index -= perSector
		next = binary.LittleEndian.Uint32(sec[perSector*4:])
	}
	return 0, false
}
