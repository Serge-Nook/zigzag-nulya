package ui

import (
	"net/url"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// bulkTab — массовые операции над всеми карточками.
func (u *App) bulkTab() fyne.CanvasObject {
	date := widget.NewEntry()
	date.SetText(time.Now().Format("02.01.2006"))
	applyDate := widget.NewButton("Применить дату ко всем карточкам", func() {
		for i := range u.project.Cards {
			u.project.Cards[i].Date = date.Text
		}
		u.reloadEditor()
		u.status("Дата применена ко всем карточкам")
	})

	sameSobol := widget.NewButton("Один пароль «СОБОЛЬ» для всех карточек", func() {
		p, err := u.gen.Password()
		if err != nil {
			u.errorf(err)
			return
		}
		for i := range u.project.Cards {
			u.project.Cards[i].SobolPass = p
		}
		u.reloadEditor()
		u.status("Единый пароль «СОБОЛЬ» установлен")
	})
	uniqueSobol := widget.NewButton("Уникальный пароль «СОБОЛЬ» для каждой карточки", func() {
		u.regenerate(true)
	})
	uniqueOS := widget.NewButton("Разные пароли О.С. для всех карточек", func() {
		u.regenerate(false)
	})
	sobolAll := widget.NewButton("Включить «СОБОЛЬ» во всех карточках", func() {
		u.setSobolAll(true)
	})
	sobolNone := widget.NewButton("Отключить «СОБОЛЬ» во всех карточках", func() {
		u.setSobolAll(false)
	})

	return container.NewVScroll(container.NewVBox(
		widget.NewLabelWithStyle("Дата", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		date, applyDate,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Пароли", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		sameSobol, uniqueSobol, uniqueOS,
		widget.NewSeparator(),
		sobolAll, sobolNone,
	))
}

func (u *App) setSobolAll(v bool) {
	for i := range u.project.Cards {
		u.project.Cards[i].SobolOn = v
	}
	u.reloadEditor()
}

func (u *App) regenerate(sobol bool) {
	for i := range u.project.Cards {
		p, err := u.gen.Password()
		if err != nil {
			u.errorf(err)
			return
		}
		if sobol {
			u.project.Cards[i].SobolPass = p
		} else {
			u.project.Cards[i].OSPassword = p
		}
	}
	u.reloadEditor()
	u.status("Пароли перегенерированы")
}

func (u *App) reloadEditor() {
	if u.current >= 0 && u.current < len(u.project.Cards) {
		u.editor.load(&u.project.Cards[u.current])
	}
	u.refresh()
}

func (u *App) settingsTab() fyne.CanvasObject {
	themes := map[string]string{"Светлая": "light", "Темная": "dark", "Системная": "system"}
	names := []string{"Светлая", "Темная", "Системная"}
	sel := widget.NewSelect(names, func(s string) {
		u.project.Settings.Theme = themes[s]
		u.applyTheme(themes[s])
	})
	for n, v := range themes {
		if v == u.project.Settings.Theme {
			sel.SetSelected(n)
		}
	}
	return container.NewVBox(
		widget.NewLabelWithStyle("Тема оформления", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		sel,
	)
}

func (u *App) aboutTab() fyne.CanvasObject {
	site, _ := url.Parse("https://sd-on.ru")
	donate, _ := url.Parse("https://sd-on.ru/donate/")
	return container.NewVBox(
		widget.NewLabelWithStyle("КОРОБ", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Версия: "+Version),
		widget.NewLabel("Автор: Сергей Владимирович Горшков"),
		container.NewHBox(widget.NewLabel("Сайт:"), widget.NewHyperlink("https://sd-on.ru", site)),
		container.NewHBox(widget.NewLabel("Пожертвования:"), widget.NewHyperlink("https://sd-on.ru/donate/", donate)),
		widget.NewLabel("Разработано на языке: Go (Golang)"),
		widget.NewLabel("Поддерживаемые платформы: Arch Linux, Debian, Astra Linux"),
	)
}

func (u *App) showAbout() {
	dialog.ShowCustom("О программе", "Закрыть", u.aboutTab(), u.win)
}
