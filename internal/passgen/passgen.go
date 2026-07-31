// Package passgen генерирует пароли с использованием криптографически
// безопасного источника случайных данных (crypto/rand).
package passgen

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"unicode"
)

// SpecialChars — допустимые специальные символы обрамления первой части пароля.
var SpecialChars = []rune{'!', '-', '=', '(', ')'}

const (
	minWordLen = 6
	maxWordLen = 10
	// MinLength — минимальная длина пароля при ручном вводе.
	MinLength = 12
	// MaxLength — максимальная длина пароля при ручном вводе.
	// Пароль складывается из 8 символов первой части и слова длиной 6–10 букв,
	// поэтому верхняя граница равна 18.
	MaxLength = 18
	// genWordMax — максимальная длина слова при автоматической генерации,
	// подобранная так, чтобы пароль укладывался в 16 символов.
	genWordMax = 8
)

// ErrExhausted возвращается, если не удалось подобрать уникальный пароль.
var ErrExhausted = errors.New("не удалось сгенерировать уникальный пароль")

// Generator создаёт уникальные пароли и запоминает уже выданные значения.
type Generator struct {
	used map[string]bool
}

// New создаёт генератор, помнящий ранее выданные пароли.
func New(used []string) *Generator {
	g := &Generator{used: map[string]bool{}}
	for _, u := range used {
		g.used[u] = true
	}
	return g
}

// Used возвращает список всех использованных паролей.
func (g *Generator) Used() []string {
	out := make([]string, 0, len(g.used))
	for k := range g.used {
		out = append(out, k)
	}
	return out
}

// Remember помечает пароль как использованный.
func (g *Generator) Remember(p string) {
	if p != "" {
		g.used[p] = true
	}
}

func randInt(n int) int {
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		panic(fmt.Sprintf("crypto/rand: %v", err))
	}
	return int(v.Int64())
}

// FirstPart генерирует первую часть пароля вида <спецсимвол><6 цифр><тот же спецсимвол>.
func FirstPart() string {
	s := SpecialChars[randInt(len(SpecialChars))]
	var digits strings.Builder
	for i := 0; i < 6; i++ {
		digits.WriteByte(byte('0' + randInt(10)))
	}
	return string(s) + digits.String() + string(s)
}

// SecondPart возвращает английское слово из 6–10 букв с заглавной первой буквой.
func SecondPart() string {
	for {
		w := words[randInt(len(words))]
		if len(w) > genWordMax {
			continue
		}
		return strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
	}
}

// Password генерирует уникальный пароль целиком.
func (g *Generator) Password() (string, error) {
	for i := 0; i < 10000; i++ {
		p := FirstPart() + SecondPart()
		if !g.used[p] {
			g.used[p] = true
			return p, nil
		}
	}
	return "", ErrExhausted
}

// UniqueFirstPart генерирует уникальную первую часть с учётом заданной второй части.
func (g *Generator) UniqueFirstPart(second string) (string, error) {
	for i := 0; i < 10000; i++ {
		p := FirstPart()
		if !g.used[p+second] {
			g.used[p+second] = true
			return p, nil
		}
	}
	return "", ErrExhausted
}

// UniqueSecondPart генерирует уникальную вторую часть с учётом заданной первой части.
func (g *Generator) UniqueSecondPart(first string) (string, error) {
	for i := 0; i < 10000; i++ {
		p := SecondPart()
		if !g.used[first+p] {
			g.used[first+p] = true
			return p, nil
		}
	}
	return "", ErrExhausted
}

// ValidateFirstPart проверяет формат первой части пароля.
func ValidateFirstPart(s string) error {
	r := []rune(s)
	if len(r) != 8 {
		return errors.New("первая часть должна состоять из 8 символов")
	}
	if r[0] != r[7] {
		return errors.New("первый и последний символы должны совпадать")
	}
	ok := false
	for _, c := range SpecialChars {
		if c == r[0] {
			ok = true
		}
	}
	if !ok {
		return errors.New("допустимые спецсимволы: ! - = ( )")
	}
	for _, c := range r[1:7] {
		if !unicode.IsDigit(c) {
			return errors.New("между спецсимволами должны быть шесть цифр")
		}
	}
	return nil
}

// ValidateSecondPart проверяет формат второй части пароля.
func ValidateSecondPart(s string) error {
	r := []rune(s)
	if len(r) < minWordLen || len(r) > maxWordLen {
		return fmt.Errorf("слово должно содержать от %d до %d букв", minWordLen, maxWordLen)
	}
	if !unicode.IsUpper(r[0]) {
		return errors.New("первая буква должна быть заглавной")
	}
	for _, c := range r {
		if c < 'A' || (c > 'Z' && c < 'a') || c > 'z' {
			return errors.New("допускаются только латинские буквы")
		}
	}
	return nil
}

// ValidatePassword проверяет пароль целиком (ручной ввод).
func ValidatePassword(s string) error {
	r := []rune(s)
	if len(r) < MinLength || len(r) > MaxLength {
		return fmt.Errorf("длина пароля должна быть от %d до %d символов", MinLength, MaxLength)
	}
	if len(r) < 9 {
		return errors.New("некорректный формат пароля")
	}
	if err := ValidateFirstPart(string(r[:8])); err != nil {
		return err
	}
	return ValidateSecondPart(string(r[8:]))
}
