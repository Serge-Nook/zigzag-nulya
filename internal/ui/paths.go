package ui

import (
	"github.com/ncruces/zenity"
)

// pickFiles opens a native multi-select file dialog.
func pickFiles(title string, filters []zenity.FileFilter) ([]string, error) {
	options := []zenity.Option{zenity.Title(title)}
	for _, filter := range filters {
		options = append(options, zenity.FileFilters{filter})
	}
	return zenity.SelectFileMultiple(options...)
}

// pickDirectory opens a native directory chooser.
func pickDirectory(title string) (string, error) {
	return zenity.SelectFile(zenity.Title(title), zenity.Directory())
}
