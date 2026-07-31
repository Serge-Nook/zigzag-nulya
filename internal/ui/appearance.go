package ui

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/Serge-Nook/korob/internal/model"
)

const rowHeight = float32(44)

// fieldRow — строка настроек поля с поддержкой перетаскивания (Drag & Drop).
type fieldRow struct {
	widget.BaseWidget
	content fyne.CanvasObject
	index   int
	offset  float32
	onMove  func(from, to int)
}

func newFieldRow(index int, content fyne.CanvasObject, onMove func(from, to int)) *fieldRow {
	r := &fieldRow{content: content, index: index, onMove: onMove}
	r.ExtendBaseWidget(r)
	return r
}

// CreateRenderer реализует интерфейс fyne.Widget.
func (r *fieldRow) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(r.content)
}

// Dragged накапливает смещение строки при перетаскивании.
func (r *fieldRow) Dragged(e *fyne.DragEvent) { r.offset += e.Dragged.DY }

// DragEnd применяет перемещение поля на новую позицию.
func (r *fieldRow) DragEnd() {
	shift := int(r.offset / rowHeight)
	if r.offset < 0 {
		shift = -int(-r.offset / rowHeight)
	}
	r.offset = 0
	if shift != 0 && r.onMove != nil {
		r.onMove(r.index, r.index+shift)
	}
}

func (u *App) appearanceTab() fyne.CanvasObject {
	list := container.NewVBox()
	logoPreview := canvas.NewImageFromImage(nil)
	logoPreview.FillMode = canvas.ImageFillContain
	logoPreview.SetMinSize(fyne.NewSize(96, 96))

	var rebuild func()
	move := func(from, to int) {
		if to < 0 {
			to = 0
		}
		if to >= len(u.project.Appearance.Fields) {
			to = len(u.project.Appearance.Fields) - 1
		}
		u.project.Appearance.Move(from, to)
		rebuild()
		u.updatePreview()
	}

	rebuild = func() {
		list.Objects = nil
		for i := range u.project.Appearance.Fields {
			i := i
			st := &u.project.Appearance.Fields[i]
			title := model.FieldTitles[st.Field]
			if title == "" {
				title = "Дата"
			}
			visible := widget.NewCheck("показывать", func(v bool) {
				st.Visible = v
				u.updatePreview()
			})
			visible.SetChecked(st.Visible)
			bold := widget.NewCheck("жирный", func(v bool) {
				st.Bold = v
				u.updatePreview()
			})
			bold.SetChecked(st.Bold)
			size := widget.NewSelect(fontSizes(), func(s string) {
				if v, err := strconv.ParseFloat(s, 64); err == nil {
					st.FontSize = v
					u.updatePreview()
				}
			})
			size.SetSelected(strconv.Itoa(int(st.FontSize)))
			up := widget.NewButtonWithIcon("", theme.MoveUpIcon(), func() { move(i, i-1) })
			down := widget.NewButtonWithIcon("", theme.MoveDownIcon(), func() { move(i, i+1) })
			row := container.NewBorder(nil, nil,
				container.NewHBox(widget.NewIcon(theme.MenuIcon()), widget.NewLabel(title)),
				container.NewHBox(size, bold, visible, up, down),
			)
			list.Add(newFieldRow(i, row, move))
		}
		list.Refresh()
	}
	rebuild()

	setLogo := func(data []byte) {
		u.project.Appearance.LogoJPEG = data
		if len(data) == 0 {
			logoPreview.Image = nil
		} else if img, err := jpeg.Decode(bytes.NewReader(data)); err == nil {
			logoPreview.Image = img
		}
		logoPreview.Refresh()
	}
	setLogo(u.project.Appearance.LogoJPEG)

	loadLogo := widget.NewButton("Загрузить логотип (JPG, квадратный)", func() {
		d := dialog.NewFileOpen(func(rc fyne.URIReadCloser, err error) {
			if err != nil || rc == nil {
				return
			}
			defer rc.Close()
			data, err := io.ReadAll(rc)
			if err != nil {
				u.errorf(err)
				return
			}
			cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
			if err != nil || format != "jpeg" {
				u.errorf(fmt.Errorf("логотип должен быть в формате JPG"))
				return
			}
			if cfg.Width != cfg.Height {
				u.errorf(fmt.Errorf("изображение должно быть квадратным (%dx%d)", cfg.Width, cfg.Height))
				return
			}
			setLogo(data)
			u.status("Логотип загружен")
		}, u.win)
		d.SetFilter(storage.NewExtensionFileFilter([]string{".jpg", ".jpeg"}))
		d.Show()
	})
	clearLogo := widget.NewButton("Удалить логотип", func() { setLogo(nil) })

	return container.NewVScroll(container.NewVBox(
		widget.NewLabelWithStyle("Порядок и оформление полей (перетаскивайте строки)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		list,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Логотип (общий для всех карточек, правый верхний угол)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewHBox(logoPreview, container.NewVBox(loadLogo, clearLogo)),
	))
}

func fontSizes() []string {
	var out []string
	for s := 8; s <= 32; s += 2 {
		out = append(out, strconv.Itoa(s))
	}
	return out
}
