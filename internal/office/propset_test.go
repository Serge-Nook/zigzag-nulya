package office

import (
	"encoding/binary"
	"testing"
	"unicode/utf16"
)

// buildPropertySet creates a single-section SummaryInformation blob.
func buildPropertySet(t *testing.T, codePage uint16, props map[uint32][]byte) []byte {
	t.Helper()
	ps := &propertySet{header: make([]byte, 48), sectionOff: 48}
	binary.LittleEndian.PutUint16(ps.header[0:2], 0xFFFE)
	binary.LittleEndian.PutUint32(ps.header[24:28], 1)
	ps.codePage = codePage
	ps.setCodePage(codePage)
	for id, raw := range props {
		ps.set(id, raw)
	}
	return ps.bytes()
}

func decodeLPSTR(t *testing.T, raw []byte) string {
	t.Helper()
	if got := binary.LittleEndian.Uint32(raw[0:4]); got != vtLPSTR {
		t.Fatalf("ожидался VT_LPSTR, получено %d", got)
	}
	size := int(binary.LittleEndian.Uint32(raw[4:8]))
	return string(raw[8 : 8+size-1])
}

func decodeLPWSTR(t *testing.T, raw []byte) string {
	t.Helper()
	if got := binary.LittleEndian.Uint32(raw[0:4]); got != vtLPWSTR {
		t.Fatalf("ожидался VT_LPWSTR, получено %d", got)
	}
	count := int(binary.LittleEndian.Uint32(raw[4:8]))
	units := make([]uint16, 0, count-1)
	for i := 0; i < count-1; i++ {
		units = append(units, binary.LittleEndian.Uint16(raw[8+i*2:10+i*2]))
	}
	return string(utf16.Decode(units))
}

func propRaw(t *testing.T, ps *propertySet, id uint32) []byte {
	t.Helper()
	for _, prop := range ps.props {
		if prop.id == id {
			return prop.raw
		}
	}
	t.Fatalf("свойство %#x не найдено", id)
	return nil
}

func TestPropertySetRoundTripASCII(t *testing.T) {
	blob := buildPropertySet(t, 1252, map[uint32][]byte{
		pidAuthor: encodeLPSTR("Old Author"),
		0x02:      encodeLPSTR("Title"),
	})

	ps, err := parsePropertySet(blob)
	if err != nil {
		t.Fatalf("parsePropertySet: %v", err)
	}
	ps.setString(pidAuthor, "Serge Nook")
	ps.setString(pidLastAuthor, "Editor")

	reparsed, err := parsePropertySet(ps.bytes())
	if err != nil {
		t.Fatalf("повторный разбор: %v", err)
	}
	if got := decodeLPSTR(t, propRaw(t, reparsed, pidAuthor)); got != "Serge Nook" {
		t.Errorf("автор = %q", got)
	}
	if got := decodeLPSTR(t, propRaw(t, reparsed, pidLastAuthor)); got != "Editor" {
		t.Errorf("редактор = %q", got)
	}
	if got := decodeLPSTR(t, propRaw(t, reparsed, 0x02)); got != "Title" {
		t.Errorf("заголовок потерян: %q", got)
	}
	if reparsed.codePage != 1252 {
		t.Errorf("кодовая страница изменилась: %d", reparsed.codePage)
	}
}

func TestPropertySetSwitchesToUTF8ForCyrillic(t *testing.T) {
	blob := buildPropertySet(t, 1252, map[uint32][]byte{pidAuthor: encodeLPSTR("Old")})
	ps, err := parsePropertySet(blob)
	if err != nil {
		t.Fatal(err)
	}
	ps.setString(pidAuthor, "Иван Иванов")

	reparsed, err := parsePropertySet(ps.bytes())
	if err != nil {
		t.Fatal(err)
	}
	if reparsed.codePage != 65001 {
		t.Fatalf("кодовая страница = %d, ожидалась 65001", reparsed.codePage)
	}
	if got := decodeLPSTR(t, propRaw(t, reparsed, pidAuthor)); got != "Иван Иванов" {
		t.Errorf("автор = %q", got)
	}
}

func TestPropertySetKeepsUnicodeEncoding(t *testing.T) {
	blob := buildPropertySet(t, 1200, map[uint32][]byte{pidAuthor: encodeLPWSTR("Old")})
	ps, err := parsePropertySet(blob)
	if err != nil {
		t.Fatal(err)
	}
	ps.setString(pidAuthor, "Иван")

	reparsed, err := parsePropertySet(ps.bytes())
	if err != nil {
		t.Fatal(err)
	}
	if got := decodeLPWSTR(t, propRaw(t, reparsed, pidAuthor)); got != "Иван" {
		t.Errorf("автор = %q", got)
	}
}

func TestPropertySetRejectsGarbage(t *testing.T) {
	if _, err := parsePropertySet([]byte("не property set")); err == nil {
		t.Fatal("ожидалась ошибка разбора")
	}
}
