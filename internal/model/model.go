// Package model описывает данные программы «КОРОБ»: карточки пользователей,
// оформление карточек и настройки приложения.
package model

import "time"

// Идентификаторы полей карточки.
const (
	FieldDepartment = "department"
	FieldWorkplace  = "workplace"
	FieldSobol      = "sobol"
	FieldOSLogin    = "os_login"
	FieldOSPassword = "os_password"
	FieldDate       = "date"
)

// FieldOrderDefault — порядок полей карточки по умолчанию.
var FieldOrderDefault = []string{
	FieldDepartment,
	FieldWorkplace,
	FieldSobol,
	FieldOSLogin,
	FieldOSPassword,
	FieldDate,
}

// FieldTitles — подписи полей карточки.
var FieldTitles = map[string]string{
	FieldDepartment: "Отдел №",
	FieldWorkplace:  "Рабочее место (АРМ) №",
	FieldSobol:      "Пароль для «СОБОЛЬ»:",
	FieldOSLogin:    "Логин О.С.:",
	FieldOSPassword: "Пароль О.С.:",
	FieldDate:       "",
}

// Card — карточка пользователя.
type Card struct {
	ID          string `json:"id"`
	Department  string `json:"department"`
	Workplace   string `json:"workplace"`
	SobolOn     bool   `json:"sobol_on"`
	SobolPass   string `json:"sobol_pass"`
	OSLogin     string `json:"os_login"`
	OSPassword  string `json:"os_password"`
	Date        string `json:"date"`
	Description string `json:"description,omitempty"`
}

// NewCard создаёт пустую карточку с текущей датой.
func NewCard(id string) Card {
	return Card{
		ID:      id,
		SobolOn: true,
		Date:    time.Now().Format("02.01.2006"),
	}
}

// Value возвращает значение поля карточки по его идентификатору.
func (c Card) Value(field string) string {
	switch field {
	case FieldDepartment:
		return c.Department
	case FieldWorkplace:
		return c.Workplace
	case FieldSobol:
		return c.SobolPass
	case FieldOSLogin:
		return c.OSLogin
	case FieldOSPassword:
		return c.OSPassword
	case FieldDate:
		return c.Date
	}
	return ""
}

// FieldStyle — оформление одного поля карточки.
type FieldStyle struct {
	Field    string  `json:"field"`
	Visible  bool    `json:"visible"`
	Bold     bool    `json:"bold"`
	FontSize float64 `json:"font_size"`
}

// Appearance — общие настройки оформления всех карточек.
type Appearance struct {
	Fields   []FieldStyle `json:"fields"`
	LogoJPEG []byte       `json:"logo_jpeg,omitempty"`
}

// DefaultAppearance возвращает оформление по умолчанию.
func DefaultAppearance() Appearance {
	a := Appearance{}
	for _, f := range FieldOrderDefault {
		a.Fields = append(a.Fields, FieldStyle{Field: f, Visible: true, FontSize: 12})
	}
	return a
}

// Style возвращает стиль поля (и признак того, что стиль найден).
func (a Appearance) Style(field string) (FieldStyle, bool) {
	for _, f := range a.Fields {
		if f.Field == field {
			return f, true
		}
	}
	return FieldStyle{}, false
}

// Move перемещает поле с позиции from на позицию to (Drag & Drop).
func (a *Appearance) Move(from, to int) {
	if from < 0 || to < 0 || from >= len(a.Fields) || to >= len(a.Fields) || from == to {
		return
	}
	f := a.Fields[from]
	a.Fields = append(a.Fields[:from], a.Fields[from+1:]...)
	rest := append([]FieldStyle{f}, a.Fields[to:]...)
	a.Fields = append(a.Fields[:to:to], rest...)
}

// Settings — настройки приложения.
type Settings struct {
	Theme string `json:"theme"` // light | dark | system
}

// Project — все данные программы (формат *.7box).
type Project struct {
	Version    int        `json:"version"`
	Cards      []Card     `json:"cards"`
	Appearance Appearance `json:"appearance"`
	Settings   Settings   `json:"settings"`
	UsedPass   []string   `json:"used_passwords"`
}

// NewProject создаёт пустой проект.
func NewProject() *Project {
	return &Project{
		Version:    1,
		Appearance: DefaultAppearance(),
		Settings:   Settings{Theme: "system"},
	}
}

// Lines возвращает текстовые строки карточки в порядке, заданном оформлением.
func (p *Project) Lines(c Card) []struct {
	Text  string
	Style FieldStyle
} {
	var out []struct {
		Text  string
		Style FieldStyle
	}
	for _, st := range p.Appearance.Fields {
		if !st.Visible {
			continue
		}
		if st.Field == FieldSobol && !c.SobolOn {
			continue
		}
		title := FieldTitles[st.Field]
		text := c.Value(st.Field)
		line := text
		if title != "" {
			line = title + " " + text
		}
		out = append(out, struct {
			Text  string
			Style FieldStyle
		}{Text: line, Style: st})
	}
	return out
}
