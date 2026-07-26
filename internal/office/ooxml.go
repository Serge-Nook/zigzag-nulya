package office

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const corePropsPath = "docProps/core.xml"

const corePropsTemplate = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:dcmitype="http://purl.org/dc/dcmitype/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"><dc:creator></dc:creator><cp:lastModifiedBy></cp:lastModifiedBy></cp:coreProperties>`

// updateOOXML rewrites docProps/core.xml inside a .docx/.xlsx/.pptx package.
func updateOOXML(path string, meta Metadata) error {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("не удалось открыть файл: %w", err)
	}
	defer reader.Close()

	tmp, err := os.CreateTemp(filepath.Dir(path), ".topor-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	writer := zip.NewWriter(tmp)
	found := false
	for _, entry := range reader.File {
		if entry.Name == corePropsPath {
			found = true
			data, err := readZipEntry(entry)
			if err != nil {
				tmp.Close()
				return err
			}
			if err := copyEntry(writer, &entry.FileHeader, patchCoreProps(data, meta)); err != nil {
				tmp.Close()
				return err
			}
			continue
		}
		data, err := readZipEntry(entry)
		if err != nil {
			tmp.Close()
			return err
		}
		if err := copyEntry(writer, &entry.FileHeader, data); err != nil {
			tmp.Close()
			return err
		}
	}
	if !found {
		header := &zip.FileHeader{Name: corePropsPath, Method: zip.Deflate}
		if err := copyEntry(writer, header, patchCoreProps([]byte(corePropsTemplate), meta)); err != nil {
			tmp.Close()
			return err
		}
	}
	if err := writer.Close(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return replaceFile(tmpName, path)
}

func readZipEntry(entry *zip.File) ([]byte, error) {
	rc, err := entry.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

func copyEntry(writer *zip.Writer, header *zip.FileHeader, data []byte) error {
	h := *header
	h.Method = zip.Deflate
	w, err := writer.CreateHeader(&h)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

var (
	creatorRe      = regexp.MustCompile(`(?s)<dc:creator>.*?</dc:creator>`)
	lastModifiedRe = regexp.MustCompile(`(?s)<cp:lastModifiedBy>.*?</cp:lastModifiedBy>`)
	selfClosingRe  = regexp.MustCompile(`<(dc:creator|cp:lastModifiedBy)\s*/>`)
)

func patchCoreProps(data []byte, meta Metadata) []byte {
	xmlText := selfClosingRe.ReplaceAllString(string(data), "<$1></$1>")
	if meta.SetAuthor {
		xmlText = replaceOrInsert(xmlText, creatorRe, "dc:creator", meta.Author)
	}
	if meta.SetLastModifiedBy {
		xmlText = replaceOrInsert(xmlText, lastModifiedRe, "cp:lastModifiedBy", meta.LastModifiedBy)
	}
	return []byte(xmlText)
}

func replaceOrInsert(xmlText string, re *regexp.Regexp, tag, value string) string {
	element := fmt.Sprintf("<%s>%s</%s>", tag, escapeXML(value), tag)
	if re.MatchString(xmlText) {
		return re.ReplaceAllLiteralString(xmlText, element)
	}
	if idx := strings.LastIndex(xmlText, "</cp:coreProperties>"); idx >= 0 {
		return xmlText[:idx] + element + xmlText[idx:]
	}
	return xmlText
}

func escapeXML(value string) string {
	var buf bytes.Buffer
	_ = xml.EscapeText(&buf, []byte(value))
	return buf.String()
}
