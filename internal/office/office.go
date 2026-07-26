// Package office edits the internal author metadata of Word and Excel files.
package office

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Metadata holds the author fields that have to be written into a document.
type Metadata struct {
	Author            string
	SetAuthor         bool
	LastModifiedBy    string
	SetLastModifiedBy bool
}

// Result reports the outcome for a single processed document.
type Result struct {
	Path string
	Err  error
}

// ErrUnsupported is returned for files that are not Word/Excel documents.
var ErrUnsupported = errors.New("неподдерживаемый формат файла")

// SupportedExtensions lists the document types the editor can handle.
var SupportedExtensions = []string{".docx", ".docm", ".doc", ".xlsx", ".xlsm", ".xls"}

// IsSupported reports whether the file extension can be processed.
func IsSupported(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, supported := range SupportedExtensions {
		if ext == supported {
			return true
		}
	}
	return false
}

// Apply writes the metadata into every given document.
func Apply(paths []string, meta Metadata) []Result {
	results := make([]Result, 0, len(paths))
	for _, path := range paths {
		results = append(results, Result{Path: path, Err: ApplyOne(path, meta)})
	}
	return results
}

// ApplyOne writes the metadata into a single document.
func ApplyOne(path string, meta Metadata) error {
	if !meta.SetAuthor && !meta.SetLastModifiedBy {
		return nil
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".docx", ".docm", ".xlsx", ".xlsm", ".pptx":
		return updateOOXML(path, meta)
	case ".doc", ".xls", ".ppt":
		return updateLegacy(path, meta)
	default:
		return fmt.Errorf("%w: %s", ErrUnsupported, filepath.Ext(path))
	}
}

// replaceFile moves src over dst, preserving the original file mode.
func replaceFile(src, dst string) error {
	if info, err := os.Stat(dst); err == nil {
		_ = os.Chmod(src, info.Mode())
	}
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
