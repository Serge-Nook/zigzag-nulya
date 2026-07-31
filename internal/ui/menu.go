package ui

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"github.com/Serge-Nook/korob/internal/model"
	"github.com/Serge-Nook/korob/internal/passgen"
	"github.com/Serge-Nook/korob/internal/printing"
	"github.com/Serge-Nook/korob/internal/store"
)

func (u *App) mainMenu() *fyne.MainMenu {
	file := fyne.NewMenu("Файл",
		fyne.NewMenuItem("Экспорт проекта (*.7box)...", u.exportProject),
		fyne.NewMenuItem("Импорт проекта (*.7box)...", u.importProject),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Экспорт в CSV...", u.exportCSV),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Печать...", u.showPrintDialog),
	)
	help := fyne.NewMenu("Справка", fyne.NewMenuItem("О программе", u.showAbout))
	return fyne.NewMainMenu(file, help)
}

func (u *App) saveDialog(ext string, save func(path string) error) {
	d := dialog.NewFileSave(func(wc fyne.URIWriteCloser, err error) {
		if err != nil || wc == nil {
			return
		}
		path := wc.URI().Path()
		_ = wc.Close()
		if !strings.HasSuffix(strings.ToLower(path), ext) {
			path += ext
		}
		if err := save(path); err != nil {
			u.errorf(err)
			return
		}
		u.status("Сохранено: %s", path)
	}, u.win)
	d.SetFileName("korob" + ext)
	d.Show()
}

func (u *App) exportProject() {
	u.project.UsedPass = u.gen.Used()
	u.saveDialog(".7box", func(path string) error { return store.SaveProject(path, u.project) })
}

func (u *App) exportCSV() {
	u.saveDialog(".csv", func(path string) error { return store.ExportCSV(path, u.project.Cards) })
}

func (u *App) importProject() {
	d := dialog.NewFileOpen(func(rc fyne.URIReadCloser, err error) {
		if err != nil || rc == nil {
			return
		}
		path := rc.URI().Path()
		_ = rc.Close()
		p, err := store.LoadProject(path)
		if err != nil {
			u.errorf(err)
			return
		}
		u.project = p
		u.gen = passgen.New(p.UsedPass)
		u.selected = map[string]bool{}
		u.current = -1
		u.applyTheme(p.Settings.Theme)
		u.win.SetContent(u.buildContent())
		u.refresh()
		u.status("Импортировано карточек: %d", len(p.Cards))
	}, u.win)
	d.SetFilter(storage.NewExtensionFileFilter([]string{".7box"}))
	d.Show()
}

func (u *App) showPrintDialog() {
	layout := widget.NewSelect([]string{"1 карточка на A4", "4 карточки на A4", "6 карточек на A4"}, nil)
	layout.SetSelectedIndex(1)
	scope := widget.NewRadioGroup([]string{"Все карточки", "Только выбранные"}, nil)
	scope.SetSelected("Все карточки")
	separators := widget.NewCheck("Разделять карточки линиями", nil)
	separators.SetChecked(true)
	printers := printing.Printers()
	printer := widget.NewSelect(printers, nil)
	if len(printers) > 0 {
		printer.SetSelectedIndex(0)
	}

	content := container.NewVBox(
		widget.NewLabel("Формат печати"), layout,
		scope, separators,
		widget.NewLabel("Принтер (односторонняя печать)"), printer,
	)

	cards := func() ([]model.Card, error) {
		list := u.project.Cards
		if scope.Selected == "Только выбранные" {
			list = u.selectedCards()
		}
		if len(list) == 0 {
			return nil, errors.New("нет карточек для печати")
		}
		return list, nil
	}
	opts := func() printing.Options {
		l := printing.Layout4
		switch layout.SelectedIndex() {
		case 0:
			l = printing.Layout1
		case 2:
			l = printing.Layout6
		}
		return printing.Options{Layout: l, Separators: separators.Checked}
	}

	d := dialog.NewCustomConfirm("Печать карточек", "Печать", "Отмена", content, func(ok bool) {
		if !ok {
			return
		}
		list, err := cards()
		if err != nil {
			u.errorf(err)
			return
		}
		path := filepath.Join(os.TempDir(), "korob-print.pdf")
		if err := printing.BuildPDF(path, u.project, list, opts()); err != nil {
			u.errorf(err)
			return
		}
		if err := printing.Print(path, printer.Selected); err != nil {
			u.errorf(err)
			return
		}
		u.status("Отправлено на печать: %d карточек", len(list))
	}, u.win)

	savePDF := widget.NewButton("Сохранить в PDF...", func() {
		list, err := cards()
		if err != nil {
			u.errorf(err)
			return
		}
		u.saveDialog(".pdf", func(path string) error {
			return printing.BuildPDF(path, u.project, list, opts())
		})
	})
	content.Add(savePDF)
	d.Resize(fyne.NewSize(460, 420))
	d.Show()
}
