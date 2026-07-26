package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// logView shows the result of the last operation.
type logView struct {
	entry *widget.Entry
}

func newLogView() *logView {
	entry := widget.NewMultiLineEntry()
	entry.Wrapping = fyne.TextWrapWord
	entry.SetMinRowsVisible(5)
	return &logView{entry: entry}
}

func (l *logView) set(text string) { l.entry.SetText(text) }

func (l *logView) content() fyne.CanvasObject {
	return container.NewVBox(widget.NewLabel("Результат:"), l.entry)
}
