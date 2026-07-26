package ui

import (
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// AuthorSite is the personal site of the program author.
const AuthorSite = "https://nookbat.ru"

func newAboutTab() fyne.CanvasObject {
	site, _ := url.Parse(AuthorSite)
	link := widget.NewHyperlink("nookbat.ru", site)

	return container.NewCenter(container.NewVBox(
		widget.NewLabelWithStyle(AppName, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Версия "+Version, fyne.TextAlignCenter, fyne.TextStyle{}),
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Автор: Serge Nook", fyne.TextAlignCenter, fyne.TextStyle{}),
		container.NewHBox(widget.NewLabel("Сайт:"), link),
	))
}
