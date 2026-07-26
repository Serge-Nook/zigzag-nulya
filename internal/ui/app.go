// Package ui builds the T0P0R desktop interface.
package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

// AppName is the displayed name of the program.
const AppName = "T0P0R"

// Version is overridden at build time with -ldflags.
var Version = "dev"

// Run starts the application event loop.
func Run() {
	application := app.NewWithID("ru.nookbat.topor")
	window := application.NewWindow(AppName)

	tabs := container.NewAppTabs(
		container.NewTabItem("Даты файлов", newDatesTab(window)),
		container.NewTabItem("Автор документов", newOfficeTab(window)),
		container.NewTabItem("Об авторе", newAboutTab()),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	window.SetContent(tabs)
	window.Resize(fyne.NewSize(760, 620))
	window.ShowAndRun()
}
