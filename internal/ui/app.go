// Package ui реализует графический интерфейс программы «КОРОБ» на Fyne v2.
package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/Serge-Nook/korob/internal/model"
	"github.com/Serge-Nook/korob/internal/passgen"
	"github.com/Serge-Nook/korob/internal/store"
)

// Version — версия программы.
const Version = "1.0"

// App — состояние графического приложения.
type App struct {
	fyne    fyne.App
	win     fyne.Window
	project *model.Project
	gen     *passgen.Generator

	selected  map[string]bool
	current   int
	list      *widget.List
	editor    *cardEditor
	preview   *widget.RichText
	statusBar *widget.Label
}

// Run запускает приложение.
func Run() {
	a := app.NewWithID("ru.nookbat.korob")
	u := &App{
		fyne:     a,
		project:  model.NewProject(),
		selected: map[string]bool{},
		current:  -1,
	}
	if path, err := store.ConfigPath(); err == nil {
		if p, err := store.LoadProject(path); err == nil {
			u.project = p
		}
	}
	u.gen = passgen.New(u.project.UsedPass)
	u.win = a.NewWindow("КОРОБ")
	u.win.Resize(fyne.NewSize(1100, 720))
	u.applyTheme(u.project.Settings.Theme)
	u.win.SetMainMenu(u.mainMenu())
	u.win.SetContent(u.buildContent())
	u.win.SetCloseIntercept(func() {
		u.autosave()
		u.win.Close()
	})
	u.refresh()
	u.win.ShowAndRun()
}

func (u *App) applyTheme(name string) {
	switch name {
	case "light":
		u.fyne.Settings().SetTheme(theme.LightTheme())
	case "dark":
		u.fyne.Settings().SetTheme(theme.DarkTheme())
	default:
		u.fyne.Settings().SetTheme(theme.DefaultTheme())
	}
}

func (u *App) buildContent() fyne.CanvasObject {
	u.statusBar = widget.NewLabel("")
	u.editor = newCardEditor(u)
	u.preview = widget.NewRichText()
	u.preview.Wrapping = fyne.TextWrapWord

	u.list = widget.NewList(
		func() int { return len(u.project.Cards) },
		func() fyne.CanvasObject {
			return container.NewBorder(nil, nil, widget.NewCheck("", nil), nil, widget.NewLabel(""))
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			c := u.project.Cards[i]
			box := o.(*fyne.Container)
			check := box.Objects[1].(*widget.Check)
			label := box.Objects[0].(*widget.Label)
			label.SetText(fmt.Sprintf("Отдел %s / АРМ %s", dash(c.Department), dash(c.Workplace)))
			check.OnChanged = nil
			check.SetChecked(u.selected[c.ID])
			id := c.ID
			check.OnChanged = func(v bool) { u.selected[id] = v }
		},
	)
	u.list.OnSelected = func(i widget.ListItemID) {
		u.current = i
		u.editor.load(&u.project.Cards[i])
		u.updatePreview()
	}

	tools := container.NewHBox(
		widget.NewButtonWithIcon("Добавить", theme.ContentAddIcon(), u.addCard),
		widget.NewButtonWithIcon("Удалить", theme.DeleteIcon(), u.deleteCard),
		widget.NewButtonWithIcon("Выбрать все", theme.CheckButtonCheckedIcon(), func() { u.selectAll(true) }),
		widget.NewButtonWithIcon("Снять выбор", theme.CheckButtonIcon(), func() { u.selectAll(false) }),
	)
	left := container.NewBorder(tools, nil, nil, nil, u.list)

	tabs := container.NewAppTabs(
		container.NewTabItem("Карточка", u.editor.content),
		container.NewTabItem("Оформление", u.appearanceTab()),
		container.NewTabItem("Массовые операции", u.bulkTab()),
		container.NewTabItem("Настройки", u.settingsTab()),
		container.NewTabItem("О программе", u.aboutTab()),
	)

	right := container.NewBorder(widget.NewLabelWithStyle("Предпросмотр карточки", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		nil, nil, nil, container.NewVScroll(u.preview))

	split := container.NewHSplit(container.NewHSplit(left, tabs), right)
	split.Offset = 0.68
	return container.NewBorder(nil, u.statusBar, nil, nil, split)
}

func dash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func (u *App) status(format string, args ...any) {
	if u.statusBar != nil {
		u.statusBar.SetText(fmt.Sprintf(format, args...))
	}
}

func (u *App) errorf(err error) {
	if err != nil {
		dialog.ShowError(err, u.win)
	}
}

func (u *App) addCard() {
	c := model.NewCard(fmt.Sprintf("card-%d", time.Now().UnixNano()))
	if len(u.project.Cards) > 0 {
		last := u.project.Cards[len(u.project.Cards)-1]
		c.Department = last.Department
		c.Date = last.Date
	}
	if p, err := u.gen.Password(); err == nil {
		c.SobolPass = p
	}
	if p, err := u.gen.Password(); err == nil {
		c.OSPassword = p
	}
	u.project.Cards = append(u.project.Cards, c)
	u.refresh()
	u.list.Select(len(u.project.Cards) - 1)
}

func (u *App) deleteCard() {
	if u.current < 0 || u.current >= len(u.project.Cards) {
		return
	}
	dialog.ShowConfirm("Удаление", "Удалить выбранную карточку?", func(ok bool) {
		if !ok {
			return
		}
		i := u.current
		delete(u.selected, u.project.Cards[i].ID)
		u.project.Cards = append(u.project.Cards[:i], u.project.Cards[i+1:]...)
		u.current = -1
		u.editor.load(nil)
		u.refresh()
	}, u.win)
}

func (u *App) selectAll(v bool) {
	for _, c := range u.project.Cards {
		u.selected[c.ID] = v
	}
	u.list.Refresh()
}

// selectedCards возвращает отмеченные карточки, либо все карточки, если ничего не отмечено.
func (u *App) selectedCards() []model.Card {
	var out []model.Card
	for _, c := range u.project.Cards {
		if u.selected[c.ID] {
			out = append(out, c)
		}
	}
	return out
}

func (u *App) refresh() {
	u.project.UsedPass = u.gen.Used()
	if u.list != nil {
		u.list.Refresh()
	}
	u.updatePreview()
	u.status("Карточек: %d", len(u.project.Cards))
}

func (u *App) updatePreview() {
	if u.preview == nil {
		return
	}
	u.preview.Segments = nil
	if u.current >= 0 && u.current < len(u.project.Cards) {
		for _, l := range u.project.Lines(u.project.Cards[u.current]) {
			size := float32(l.Style.FontSize)
			u.preview.Segments = append(u.preview.Segments, &widget.TextSegment{
				Text: l.Text + "\n",
				Style: widget.RichTextStyle{
					TextStyle: fyne.TextStyle{Bold: l.Style.Bold},
					SizeName:  sizeName(size),
				},
			})
		}
	}
	u.preview.Refresh()
}

func sizeName(size float32) fyne.ThemeSizeName {
	switch {
	case size >= 20:
		return theme.SizeNameHeadingText
	case size >= 16:
		return theme.SizeNameSubHeadingText
	case size <= 10:
		return theme.SizeNameCaptionText
	default:
		return theme.SizeNameText
	}
}

func (u *App) autosave() {
	path, err := store.ConfigPath()
	if err != nil {
		return
	}
	u.project.UsedPass = u.gen.Used()
	_ = store.SaveProject(path, u.project)
}
