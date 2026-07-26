package ui

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/Serge-Nook/zigzag-nulya/internal/filedates"
)

// dateLayouts lists the accepted input formats for the date fields.
var dateLayouts = []string{
	"02.01.2006 15:04:05",
	"02.01.2006 15:04",
	"02.01.2006",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	"2006-01-02",
}

const dateHint = "дд.мм.гггг чч:мм:сс"

func parseDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	for _, layout := range dateLayouts {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("не удалось разобрать дату %q, ожидается формат %s", value, dateHint)
}

func newDatesTab(window fyne.Window) fyne.CanvasObject {
	files := newFileList()
	log := newLogView()

	created := widget.NewEntry()
	created.SetPlaceHolder(dateHint)
	createdCheck := widget.NewCheck("Изменить дату создания", nil)
	createdCheck.SetChecked(true)

	modified := widget.NewEntry()
	modified.SetPlaceHolder(dateHint)
	modifiedCheck := widget.NewCheck("Изменить дату изменения", nil)
	modifiedCheck.SetChecked(true)

	now := widget.NewButton("Подставить текущее время", func() {
		stamp := time.Now().Format(dateLayouts[0])
		created.SetText(stamp)
		modified.SetText(stamp)
	})

	addFiles := widget.NewButton("Добавить файлы", func() {
		selected, err := pickFiles("Выберите файлы", nil)
		if err == nil {
			files.add(selected...)
		}
	})
	addFolder := widget.NewButton("Добавить папку", func() {
		dir, err := pickDirectory("Выберите папку")
		if err == nil && dir != "" {
			files.add(dir)
		}
	})
	clear := widget.NewButton("Очистить список", files.clear)
	files.list.OnSelected = func(id widget.ListItemID) {
		files.removeSelected(id)
		files.list.UnselectAll()
	}

	apply := widget.NewButton("Применить", func() {
		if len(files.paths) == 0 {
			log.set("Список файлов пуст.")
			return
		}
		opts := filedates.Options{SetCreated: createdCheck.Checked, SetModified: modifiedCheck.Checked}
		if !opts.SetCreated && !opts.SetModified {
			log.set("Не выбрано ни одно поле для изменения.")
			return
		}
		if opts.SetCreated {
			value, err := parseDate(created.Text)
			if err != nil {
				log.set(err.Error())
				return
			}
			opts.Created = value
		}
		if opts.SetModified {
			value, err := parseDate(modified.Text)
			if err != nil {
				log.set(err.Error())
				return
			}
			opts.Modified = value
		}
		results := filedates.Apply(files.paths, opts)
		log.set(formatDateResults(results))
	})
	apply.Importance = widget.HighImportance

	form := container.NewVBox(
		widget.NewLabelWithStyle("Даты файлов", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		createdCheck, created,
		modifiedCheck, modified,
		now,
		container.NewGridWithColumns(3, addFiles, addFolder, clear),
		apply,
	)
	if !filedates.CreationTimeSupported {
		form.Add(widget.NewLabel("Внимание: дата создания изменяется только в Windows."))
	}

	return container.NewBorder(form, log.content(), nil, nil, files.content())
}

func formatDateResults(results []filedates.Result) string {
	var failed []string
	ok := 0
	for _, result := range results {
		if result.Err != nil {
			failed = append(failed, fmt.Sprintf("%s — %v", result.Path, result.Err))
			continue
		}
		ok++
	}
	report := fmt.Sprintf("Обработано успешно: %d, с ошибками: %d", ok, len(failed))
	if len(failed) > 0 {
		report += "\n" + strings.Join(failed, "\n")
	}
	return report
}
