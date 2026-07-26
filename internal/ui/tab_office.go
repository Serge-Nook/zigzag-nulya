package ui

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/ncruces/zenity"

	"github.com/Serge-Nook/zigzag-nulya/internal/office"
)

func newOfficeTab(window fyne.Window) fyne.CanvasObject {
	files := newFileList()
	log := newLogView()

	author := widget.NewEntry()
	author.SetPlaceHolder("Иван Иванов")
	authorCheck := widget.NewCheck("Изменить автора (создателя)", nil)
	authorCheck.SetChecked(true)

	editor := widget.NewEntry()
	editor.SetPlaceHolder("Иван Иванов")
	editorCheck := widget.NewCheck("Изменить последнего редактора", nil)
	editorCheck.SetChecked(true)

	filter := zenity.FileFilter{
		Name:     "Документы Word и Excel",
		Patterns: []string{"*.docx", "*.docm", "*.doc", "*.xlsx", "*.xlsm", "*.xls"},
	}
	addFiles := widget.NewButton("Добавить файлы", func() {
		selected, err := pickFiles("Выберите документы", []zenity.FileFilter{filter})
		if err == nil {
			files.add(selected...)
		}
	})
	addFolder := widget.NewButton("Добавить папку", func() {
		dir, err := pickDirectory("Выберите папку с документами")
		if err != nil || dir == "" {
			return
		}
		files.add(expandDocuments(dir)...)
	})
	clear := widget.NewButton("Очистить список", files.clear)
	files.list.OnSelected = func(id widget.ListItemID) {
		files.removeSelected(id)
		files.list.UnselectAll()
	}

	apply := widget.NewButton("Применить", func() {
		if len(files.paths) == 0 {
			log.set("Список документов пуст.")
			return
		}
		meta := office.Metadata{
			Author:            strings.TrimSpace(author.Text),
			SetAuthor:         authorCheck.Checked,
			LastModifiedBy:    strings.TrimSpace(editor.Text),
			SetLastModifiedBy: editorCheck.Checked,
		}
		if !meta.SetAuthor && !meta.SetLastModifiedBy {
			log.set("Не выбрано ни одно поле для изменения.")
			return
		}
		log.set(formatOfficeResults(office.Apply(files.paths, meta)))
	})
	apply.Importance = widget.HighImportance

	form := container.NewVBox(
		widget.NewLabelWithStyle("Автор документов Word и Excel", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		authorCheck, author,
		editorCheck, editor,
		container.NewGridWithColumns(3, addFiles, addFolder, clear),
		apply,
		widget.NewLabel("Поддерживаются: "+strings.Join(office.SupportedExtensions, ", ")),
	)
	return container.NewBorder(form, log.content(), nil, nil, files.content())
}

// expandDocuments collects supported documents from a folder recursively.
func expandDocuments(dir string) []string {
	var documents []string
	_ = filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !office.IsSupported(path) {
			return nil
		}
		documents = append(documents, path)
		return nil
	})
	return documents
}

func formatOfficeResults(results []office.Result) string {
	var failed []string
	ok := 0
	for _, result := range results {
		if result.Err != nil {
			failed = append(failed, fmt.Sprintf("%s — %v", result.Path, result.Err))
			continue
		}
		ok++
	}
	report := fmt.Sprintf("Изменено документов: %d, с ошибками: %d", ok, len(failed))
	if len(failed) > 0 {
		report += "\n" + strings.Join(failed, "\n")
	}
	return report
}
