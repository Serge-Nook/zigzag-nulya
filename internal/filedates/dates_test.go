package filedates

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestApplySetsModificationTimeRecursively(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "sub")
	if err := os.Mkdir(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	deep := filepath.Join(nested, "file.txt")
	if err := os.WriteFile(deep, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	want := time.Date(2001, 2, 3, 4, 5, 6, 0, time.Local)
	results := Apply([]string{root}, Options{Modified: want, SetModified: true})
	if len(results) < 3 {
		t.Fatalf("ожидалась обработка папки и вложенных файлов, получено %d", len(results))
	}
	for _, result := range results {
		if result.Err != nil {
			t.Fatalf("%s: %v", result.Path, result.Err)
		}
	}

	info, err := os.Stat(deep)
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().Equal(want) {
		t.Errorf("дата изменения = %v, ожидалась %v", info.ModTime(), want)
	}
}

func TestExpandDeduplicatesPaths(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if got := len(Expand([]string{dir, file, file})); got != 2 {
		t.Errorf("получено %d путей, ожидалось 2", got)
	}
}
