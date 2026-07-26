package office

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Serge-Nook/zigzag-nulya/internal/office/cfb"
)

// copyFixture puts a copy of a testdata document into a temporary directory.
func copyFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func readSummary(t *testing.T, path string) *propertySet {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	container, err := cfb.Open(data)
	if err != nil {
		t.Fatal(err)
	}
	stream, err := container.ReadStream(summaryStream)
	if err != nil {
		t.Fatal(err)
	}
	props, err := parsePropertySet(stream)
	if err != nil {
		t.Fatal(err)
	}
	return props
}

func TestApplyOneLegacyDocuments(t *testing.T) {
	for _, name := range []string{"sample.doc", "sample.xls"} {
		t.Run(name, func(t *testing.T) {
			path := copyFixture(t, name)
			before, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}

			meta := Metadata{Author: "Serge Nook", SetAuthor: true, LastModifiedBy: "Иван Редактор", SetLastModifiedBy: true}
			if err := ApplyOne(path, meta); err != nil {
				t.Fatalf("ApplyOne: %v", err)
			}

			props := readSummary(t, path)
			if got := decodeLPSTR(t, propRaw(t, props, pidAuthor)); got != "Serge Nook" {
				t.Errorf("автор = %q", got)
			}
			if got := decodeLPSTR(t, propRaw(t, props, pidLastAuthor)); got != "Иван Редактор" {
				t.Errorf("последний редактор = %q", got)
			}

			after, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			if after.Size() < before.Size() {
				t.Errorf("документ уменьшился: было %d, стало %d", before.Size(), after.Size())
			}
		})
	}
}

func TestApplyOneLegacyKeepsOtherStreams(t *testing.T) {
	path := copyFixture(t, "sample.doc")
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	before, err := cfb.Open(original)
	if err != nil {
		t.Fatal(err)
	}
	wordBefore, err := before.ReadStream("WordDocument")
	if err != nil {
		t.Fatal(err)
	}

	if err := ApplyOne(path, Metadata{Author: "X", SetAuthor: true}); err != nil {
		t.Fatal(err)
	}

	modified, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	after, err := cfb.Open(modified)
	if err != nil {
		t.Fatal(err)
	}
	wordAfter, err := after.ReadStream("WordDocument")
	if err != nil {
		t.Fatal(err)
	}
	if string(wordBefore) != string(wordAfter) {
		t.Error("содержимое документа изменилось")
	}
}
