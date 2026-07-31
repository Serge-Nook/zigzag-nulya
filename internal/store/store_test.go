package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Serge-Nook/korob/internal/model"
)

func sample() *model.Project {
	p := model.NewProject()
	c := model.NewCard("c1")
	c.Department = "9"
	c.Workplace = "1"
	c.SobolPass = "=127593=Rabbit"
	c.OSLogin = "otd7_1"
	c.OSPassword = "!238191!Numbat"
	c.Date = "23.07.2026"
	p.Cards = append(p.Cards, c)
	p.UsedPass = []string{c.SobolPass, c.OSPassword}
	return p
}

func TestProjectRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.7box")
	p := sample()
	if err := SaveProject(path, p); err != nil {
		t.Fatal(err)
	}
	got, err := LoadProject(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Cards) != 1 || got.Cards[0].OSPassword != "!238191!Numbat" {
		t.Fatalf("данные не восстановлены: %+v", got.Cards)
	}
	if len(got.Appearance.Fields) != len(p.Appearance.Fields) {
		t.Fatal("оформление не восстановлено")
	}
}

func TestLoadProjectRejectsForeignFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.7box")
	if err := os.WriteFile(path, []byte("not a 7box file"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadProject(path); err == nil {
		t.Fatal("ожидалась ошибка формата")
	}
}

func TestExportCSV(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cards.csv")
	if err := ExportCSV(path, sample().Cards); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("ожидались заголовок и одна строка, получено %d", len(lines))
	}
	if !strings.Contains(lines[1], "otd7_1") {
		t.Fatalf("строка карточки некорректна: %s", lines[1])
	}
}
