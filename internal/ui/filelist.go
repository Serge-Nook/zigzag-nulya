package ui

import (
	"path/filepath"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// fileList is a reusable widget holding the paths selected by the user.
type fileList struct {
	paths   []string
	list    *widget.List
	counter *widget.Label
}

func newFileList() *fileList {
	fl := &fileList{counter: widget.NewLabel("Выбрано: 0")}
	fl.list = widget.NewList(
		func() int { return len(fl.paths) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(fl.paths[id])
		},
	)
	return fl
}

func (fl *fileList) add(paths ...string) {
	existing := make(map[string]bool, len(fl.paths))
	for _, path := range fl.paths {
		existing[path] = true
	}
	for _, path := range paths {
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
		if !existing[path] {
			existing[path] = true
			fl.paths = append(fl.paths, path)
		}
	}
	fl.refresh()
}

func (fl *fileList) removeSelected(id widget.ListItemID) {
	if id < 0 || id >= len(fl.paths) {
		return
	}
	fl.paths = append(fl.paths[:id], fl.paths[id+1:]...)
	fl.refresh()
}

func (fl *fileList) clear() {
	fl.paths = nil
	fl.refresh()
}

func (fl *fileList) refresh() {
	fl.counter.SetText("Выбрано: " + strconv.Itoa(len(fl.paths)))
	fl.list.Refresh()
}

func (fl *fileList) content() fyne.CanvasObject {
	return container.NewBorder(nil, fl.counter, nil, nil, fl.list)
}
