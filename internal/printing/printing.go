// Package printing формирует PDF-файлы с карточками и отправляет их на печать.
package printing

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Serge-Nook/korob/assets"
	"github.com/Serge-Nook/korob/internal/model"
	"github.com/go-pdf/fpdf"
)

// Layout — количество карточек на листе A4.
type Layout int

// Поддерживаемые варианты раскладки.
const (
	Layout1 Layout = 1
	Layout4 Layout = 4
	Layout6 Layout = 6
)

// Options — параметры печати.
type Options struct {
	Layout     Layout
	Separators bool
}

const (
	pageW  = 210.0
	pageH  = 297.0
	margin = 10.0
)

func grid(l Layout) (cols, rows int) {
	switch l {
	case Layout4:
		return 2, 2
	case Layout6:
		return 2, 3
	default:
		return 1, 1
	}
}

// BuildPDF формирует PDF с карточками и возвращает путь к созданному файлу.
func BuildPDF(path string, p *model.Project, cards []model.Card, opt Options) error {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddUTF8FontFromBytes("DejaVu", "", assets.FontRegular)
	pdf.AddUTF8FontFromBytes("DejaVu", "B", assets.FontBold)
	pdf.SetAutoPageBreak(false, 0)

	var logo string
	if len(p.Appearance.LogoJPEG) > 0 {
		logo = "logo"
		pdf.RegisterImageOptionsReader(logo, fpdf.ImageOptions{ImageType: "JPG"}, bytes.NewReader(p.Appearance.LogoJPEG))
	}

	cols, rows := grid(opt.Layout)
	perPage := cols * rows
	cellW := (pageW - 2*margin) / float64(cols)
	cellH := (pageH - 2*margin) / float64(rows)

	for i, c := range cards {
		if i%perPage == 0 {
			pdf.AddPage()
			if opt.Separators {
				drawSeparators(pdf, cols, rows, cellW, cellH)
			}
		}
		idx := i % perPage
		x := margin + float64(idx%cols)*cellW
		y := margin + float64(idx/cols)*cellH
		drawCard(pdf, p, c, x, y, cellW, cellH, logo)
	}
	if len(cards) == 0 {
		pdf.AddPage()
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return pdf.OutputFileAndClose(path)
}

func drawSeparators(pdf *fpdf.Fpdf, cols, rows int, cellW, cellH float64) {
	pdf.SetDrawColor(150, 150, 150)
	pdf.SetLineWidth(0.2)
	for c := 1; c < cols; c++ {
		x := margin + float64(c)*cellW
		pdf.Line(x, margin, x, pageH-margin)
	}
	for r := 1; r < rows; r++ {
		y := margin + float64(r)*cellH
		pdf.Line(margin, y, pageW-margin, y)
	}
}

func drawCard(pdf *fpdf.Fpdf, p *model.Project, c model.Card, x, y, w, h float64, logo string) {
	pad := 6.0
	textW := w - 2*pad
	if logo != "" {
		side := minf(20, h/4)
		pdf.ImageOptions(logo, x+w-pad-side, y+pad, side, side, false, fpdf.ImageOptions{ImageType: "JPG"}, 0, "")
		textW -= side + 2
	}
	cy := y + pad
	for _, line := range p.Lines(c) {
		style := ""
		if line.Style.Bold {
			style = "B"
		}
		size := line.Style.FontSize
		if size <= 0 {
			size = 12
		}
		pdf.SetFont("DejaVu", style, size)
		lh := size * 0.45
		pdf.SetXY(x+pad, cy)
		pdf.MultiCell(textW, lh, line.Text, "", "L", false)
		cy = pdf.GetY() + 1
		if cy > y+h-pad {
			break
		}
	}
}

func minf(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// Print отправляет PDF-файл на печать через системную команду lp (односторонняя печать).
func Print(pdfPath, printer string) error {
	lp, err := exec.LookPath("lp")
	if err != nil {
		return fmt.Errorf("команда печати lp не найдена: установите CUPS или сохраните PDF")
	}
	args := []string{"-o", "sides=one-sided"}
	if printer != "" {
		args = append(args, "-d", printer)
	}
	args = append(args, pdfPath)
	out, err := exec.Command(lp, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("печать не выполнена: %s", string(out))
	}
	return nil
}

// Printers возвращает список доступных принтеров системы.
func Printers() []string {
	out, err := exec.Command("lpstat", "-a").Output()
	if err != nil {
		return nil
	}
	var names []string
	for _, l := range bytes.Split(out, []byte("\n")) {
		f := bytes.Fields(l)
		if len(f) > 0 {
			names = append(names, string(f[0]))
		}
	}
	return names
}
