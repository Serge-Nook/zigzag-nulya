package passgen

import "testing"

func TestFirstPartFormat(t *testing.T) {
	for i := 0; i < 200; i++ {
		if err := ValidateFirstPart(FirstPart()); err != nil {
			t.Fatalf("некорректная первая часть: %v", err)
		}
	}
}

func TestSecondPartFormat(t *testing.T) {
	for i := 0; i < 200; i++ {
		if err := ValidateSecondPart(SecondPart()); err != nil {
			t.Fatalf("некорректная вторая часть: %v", err)
		}
	}
}

func TestPasswordsUnique(t *testing.T) {
	g := New(nil)
	seen := map[string]bool{}
	for i := 0; i < 500; i++ {
		p, err := g.Password()
		if err != nil {
			t.Fatal(err)
		}
		if seen[p] {
			t.Fatalf("повтор пароля: %s", p)
		}
		seen[p] = true
		if err := ValidatePassword(p); err != nil {
			t.Fatalf("%s: %v", p, err)
		}
	}
}

func TestRememberedPasswordsNotRepeated(t *testing.T) {
	g := New(nil)
	p, _ := g.Password()
	g2 := New(g.Used())
	for i := 0; i < 100; i++ {
		n, _ := g2.Password()
		if n == p {
			t.Fatal("сгенерирован ранее использованный пароль")
		}
	}
}

func TestValidatePassword(t *testing.T) {
	cases := map[string]bool{
		"=127593=Rabbit":        true,
		"!238191!Numbat":        true,
		"!238191)Numbat":        false,
		"#238191#Numbat":        false,
		"!23819a!Numbat":        false,
		"!238191!numbat":        false,
		"!238191!Num":           false,
		"!238191!Numbatnumbat":  false,
		"!238191!Numbatnumbats": false,
	}
	for in, want := range cases {
		got := ValidatePassword(in) == nil
		if got != want {
			t.Errorf("%s: получено %v, ожидалось %v", in, got, want)
		}
	}
}
