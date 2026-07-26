package office

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeDocx(t *testing.T, core string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "doc.docx")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	zw := zip.NewWriter(file)
	entries := map[string]string{
		"word/document.xml": "<w:document/>",
	}
	if core != "" {
		entries[corePropsPath] = core
	}
	for name, content := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func readEntry(t *testing.T, path, name string) string {
	t.Helper()
	reader, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	for _, entry := range reader.File {
		if entry.Name != name {
			continue
		}
		data, err := readZipEntry(entry)
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	t.Fatalf("в архиве нет %s", name)
	return ""
}

func TestApplyOneDocxReplacesAuthors(t *testing.T) {
	path := writeDocx(t, `<?xml version="1.0"?><cp:coreProperties xmlns:cp="c" xmlns:dc="d"><dc:creator>Старый</dc:creator><cp:lastModifiedBy/><dc:title>t</dc:title></cp:coreProperties>`)

	meta := Metadata{Author: "Serge Nook", SetAuthor: true, LastModifiedBy: "Иван & Ко", SetLastModifiedBy: true}
	if err := ApplyOne(path, meta); err != nil {
		t.Fatalf("ApplyOne: %v", err)
	}

	core := readEntry(t, path, corePropsPath)
	if !strings.Contains(core, "<dc:creator>Serge Nook</dc:creator>") {
		t.Errorf("автор не изменён: %s", core)
	}
	if !strings.Contains(core, "<cp:lastModifiedBy>Иван &amp; Ко</cp:lastModifiedBy>") {
		t.Errorf("редактор не изменён: %s", core)
	}
	if !strings.Contains(core, "<dc:title>t</dc:title>") {
		t.Errorf("остальные свойства потеряны: %s", core)
	}
	if body := readEntry(t, path, "word/document.xml"); body != "<w:document/>" {
		t.Errorf("содержимое документа повреждено: %s", body)
	}
}

func TestApplyOneDocxCreatesCoreProps(t *testing.T) {
	path := writeDocx(t, "")

	if err := ApplyOne(path, Metadata{Author: "Serge", SetAuthor: true}); err != nil {
		t.Fatalf("ApplyOne: %v", err)
	}
	if core := readEntry(t, path, corePropsPath); !strings.Contains(core, "<dc:creator>Serge</dc:creator>") {
		t.Errorf("core.xml не создан корректно: %s", core)
	}
}

func TestApplyOneRejectsUnknownExtension(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ApplyOne(path, Metadata{Author: "x", SetAuthor: true}); err == nil {
		t.Fatal("ожидалась ошибка для неподдерживаемого формата")
	}
}
