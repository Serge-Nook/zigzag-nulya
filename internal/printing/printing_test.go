package printing

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Serge-Nook/korob/internal/model"
)

func TestBuildPDFLayouts(t *testing.T) {
	p := model.NewProject()
	for i := 0; i < 7; i++ {
		c := model.NewCard("c")
		c.Department = "9"
		c.Workplace = "1"
		c.SobolPass = "=127593=Rabbit"
		c.OSLogin = "otd7_1"
		c.OSPassword = "!238191!Numbat"
		p.Cards = append(p.Cards, c)
	}
	for _, l := range []Layout{Layout1, Layout4, Layout6} {
		path := filepath.Join(t.TempDir(), "out.pdf")
		if err := BuildPDF(path, p, p.Cards, Options{Layout: l, Separators: true}); err != nil {
			t.Fatalf("раскладка %d: %v", l, err)
		}
		st, err := os.Stat(path)
		if err != nil || st.Size() == 0 {
			t.Fatalf("раскладка %d: пустой PDF (%v)", l, err)
		}
	}
}
