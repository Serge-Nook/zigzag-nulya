// Package store отвечает за сохранение и загрузку данных программы:
// собственный формат *.7box и экспорт в CSV.
package store

import (
	"compress/gzip"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Serge-Nook/korob/internal/model"
)

const magic = "7BOX"

// SaveProject записывает проект в файл формата *.7box (GZIP + JSON).
func SaveProject(path string, p *model.Project) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(magic); err != nil {
		return err
	}
	zw := gzip.NewWriter(f)
	enc := json.NewEncoder(zw)
	enc.SetIndent("", "  ")
	if err := enc.Encode(p); err != nil {
		return err
	}
	return zw.Close()
}

// LoadProject читает проект из файла формата *.7box.
func LoadProject(path string) (*model.Project, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	head := make([]byte, len(magic))
	if _, err := io.ReadFull(f, head); err != nil {
		return nil, err
	}
	if string(head) != magic {
		return nil, fmt.Errorf("%s: неверный формат файла", filepath.Base(path))
	}
	zr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	p := model.NewProject()
	if err := json.NewDecoder(zr).Decode(p); err != nil {
		return nil, err
	}
	if len(p.Appearance.Fields) == 0 {
		p.Appearance = model.DefaultAppearance()
	}
	return p, nil
}

// ExportCSV выгружает карточки в CSV: одна карточка — одна строка.
func ExportCSV(path string, cards []model.Card) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString("\ufeff"); err != nil { // BOM для корректной кириллицы
		return err
	}
	w := csv.NewWriter(f)
	w.Comma = ';'
	header := []string{"Отдел №", "Рабочее место (АРМ) №", "Пароль для «СОБОЛЬ»", "Логин О.С.", "Пароль О.С.", "Дата"}
	if err := w.Write(header); err != nil {
		return err
	}
	for _, c := range cards {
		sobol := c.SobolPass
		if !c.SobolOn {
			sobol = ""
		}
		if err := w.Write([]string{c.Department, c.Workplace, sobol, c.OSLogin, c.OSPassword, c.Date}); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

// ConfigPath возвращает путь к файлу автосохранения текущей базы.
func ConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir = filepath.Join(dir, "korob")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "korob.7box"), nil
}
