package ui

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/Serge-Nook/korob/internal/model"
	"github.com/Serge-Nook/korob/internal/passgen"
)

var (
	reCyrLatDigitDot = regexp.MustCompile(`^[A-Za-zА-Яа-яЁё0-9.]*$`)
	reLatDigitDot    = regexp.MustCompile(`^[A-Za-z0-9._-]*$`)
	reDate           = regexp.MustCompile(`^\d{2}\.\d{2}\.\d{4}$`)
)

// passwordEditor — редактор пароля, состоящего из двух частей.
type passwordEditor struct {
	first, second *widget.Entry
	genFirst      *widget.Check
	genSecond     *widget.Check
	box           fyne.CanvasObject
	onChange      func(value string)
}

func newPasswordEditor(u *App, title string) *passwordEditor {
	p := &passwordEditor{
		first:     widget.NewEntry(),
		second:    widget.NewEntry(),
		genFirst:  widget.NewCheck("генерировать", nil),
		genSecond: widget.NewCheck("генерировать", nil),
	}
	p.first.SetPlaceHolder("!834726!")
	p.second.SetPlaceHolder("Numbat")
	p.first.Validator = func(s string) error { return passgen.ValidateFirstPart(s) }
	p.second.Validator = func(s string) error { return passgen.ValidateSecondPart(s) }
	p.genFirst.SetChecked(true)
	p.genSecond.SetChecked(true)

	notify := func(string) {
		if p.onChange != nil {
			p.onChange(p.value())
		}
	}
	p.first.OnChanged = notify
	p.second.OnChanged = notify

	regen := widget.NewButton("Сгенерировать", func() {
		p.generate(u)
		notify("")
	})
	form := container.NewVBox(
		widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewBorder(nil, nil, widget.NewLabel("Часть 1"), p.genFirst, p.first),
		container.NewBorder(nil, nil, widget.NewLabel("Часть 2"), p.genSecond, p.second),
		regen,
	)
	p.box = form
	return p
}

func (p *passwordEditor) value() string { return p.first.Text + p.second.Text }

func (p *passwordEditor) set(v string) {
	r := []rune(v)
	if len(r) >= 8 {
		p.first.SetText(string(r[:8]))
		p.second.SetText(string(r[8:]))
		return
	}
	p.first.SetText("")
	p.second.SetText("")
}

// generate обновляет части пароля, отмеченные для автоматической генерации.
func (p *passwordEditor) generate(u *App) {
	first, second := p.first.Text, p.second.Text
	if p.genSecond.Checked {
		if s, err := u.gen.UniqueSecondPart(first); err == nil {
			second = s
		}
	}
	if p.genFirst.Checked {
		if s, err := u.gen.UniqueFirstPart(second); err == nil {
			first = s
		}
	}
	p.first.SetText(first)
	p.second.SetText(second)
	u.gen.Remember(first + second)
}

type cardEditor struct {
	u       *App
	card    *model.Card
	loading bool

	department *widget.Entry
	workplace  *widget.Entry
	sobolOn    *widget.Check
	sobol      *passwordEditor
	osLogin    *widget.Entry
	osPass     *passwordEditor
	date       *widget.Entry
	content    fyne.CanvasObject
}

func newCardEditor(u *App) *cardEditor {
	e := &cardEditor{u: u}
	e.department = widget.NewEntry()
	e.workplace = widget.NewEntry()
	e.osLogin = widget.NewEntry()
	e.date = widget.NewEntry()
	e.sobolOn = widget.NewCheck("Использовать пароль «СОБОЛЬ»", nil)
	e.sobol = newPasswordEditor(u, "Пароль для «СОБОЛЬ»")
	e.osPass = newPasswordEditor(u, "Пароль О.С.")

	e.department.Validator = validator(reCyrLatDigitDot, "допустимы русские и латинские буквы, цифры и точки")
	e.workplace.Validator = validator(reCyrLatDigitDot, "допустимы русские и латинские буквы, цифры и точки")
	e.osLogin.Validator = validator(reLatDigitDot, "допустимы латинские буквы, цифры и точки")
	e.date.Validator = func(s string) error {
		if !reDate.MatchString(s) {
			return errors.New("формат даты: ДД.ММ.ГГГГ")
		}
		if _, err := time.Parse("02.01.2006", s); err != nil {
			return errors.New("несуществующая дата")
		}
		return nil
	}

	e.department.OnChanged = func(s string) { e.apply(func(c *model.Card) { c.Department = s }) }
	e.workplace.OnChanged = func(s string) { e.apply(func(c *model.Card) { c.Workplace = s }) }
	e.osLogin.OnChanged = func(s string) { e.apply(func(c *model.Card) { c.OSLogin = strings.TrimSpace(s) }) }
	e.date.OnChanged = func(s string) { e.apply(func(c *model.Card) { c.Date = s }) }
	e.sobolOn.OnChanged = func(v bool) {
		e.apply(func(c *model.Card) { c.SobolOn = v })
		e.updateSobolState(v)
	}
	e.sobol.onChange = func(v string) { e.apply(func(c *model.Card) { c.SobolPass = v }) }
	e.osPass.onChange = func(v string) { e.apply(func(c *model.Card) { c.OSPassword = v }) }

	form := widget.NewForm(
		widget.NewFormItem("Отдел №", e.department),
		widget.NewFormItem("Рабочее место (АРМ) №", e.workplace),
		widget.NewFormItem("Логин О.С.", e.osLogin),
		widget.NewFormItem("Дата", e.date),
	)
	e.content = container.NewVScroll(container.NewVBox(
		form,
		widget.NewSeparator(),
		e.sobolOn,
		e.sobol.box,
		widget.NewSeparator(),
		e.osPass.box,
	))
	e.load(nil)
	return e
}

func validator(re *regexp.Regexp, msg string) func(string) error {
	return func(s string) error {
		if !re.MatchString(s) {
			return errors.New(msg)
		}
		return nil
	}
}

func (e *cardEditor) updateSobolState(on bool) {
	if on {
		e.sobol.first.Enable()
		e.sobol.second.Enable()
		return
	}
	e.sobol.first.Disable()
	e.sobol.second.Disable()
}

func (e *cardEditor) apply(fn func(*model.Card)) {
	if e.loading || e.card == nil {
		return
	}
	fn(e.card)
	e.u.list.Refresh()
	e.u.updatePreview()
}

func (e *cardEditor) load(c *model.Card) {
	e.loading = true
	defer func() { e.loading = false }()
	e.card = c
	if c == nil {
		e.department.SetText("")
		e.workplace.SetText("")
		e.osLogin.SetText("")
		e.date.SetText("")
		e.sobol.set("")
		e.osPass.set("")
		e.sobolOn.SetChecked(false)
		return
	}
	e.department.SetText(c.Department)
	e.workplace.SetText(c.Workplace)
	e.osLogin.SetText(c.OSLogin)
	e.date.SetText(c.Date)
	e.sobolOn.SetChecked(c.SobolOn)
	e.sobol.set(c.SobolPass)
	e.osPass.set(c.OSPassword)
	e.updateSobolState(c.SobolOn)
}
