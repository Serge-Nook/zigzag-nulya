package model

import (
	"strings"
	"testing"
)

func TestMove(t *testing.T) {
	a := DefaultAppearance()
	a.Move(0, 2)
	want := []string{FieldWorkplace, FieldSobol, FieldDepartment, FieldOSLogin, FieldOSPassword, FieldDate}
	for i, f := range want {
		if a.Fields[i].Field != f {
			t.Fatalf("позиция %d: получено %s, ожидалось %s", i, a.Fields[i].Field, f)
		}
	}
	a.Move(2, 0)
	if a.Fields[0].Field != FieldDepartment {
		t.Fatalf("перемещение назад не выполнено: %s", a.Fields[0].Field)
	}
}

func TestLinesRespectVisibilityAndSobol(t *testing.T) {
	p := NewProject()
	c := NewCard("c1")
	c.Department = "9"
	c.SobolOn = false
	lines := p.Lines(c)
	for _, l := range lines {
		if strings.Contains(l.Text, "СОБОЛЬ") {
			t.Fatal("отключённое поле СОБОЛЬ попало в карточку")
		}
	}
	p.Appearance.Fields[0].Visible = false
	if len(p.Lines(c)) != len(lines)-1 {
		t.Fatal("скрытое поле отображается")
	}
}
