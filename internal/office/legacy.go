package office

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Serge-Nook/zigzag-nulya/internal/office/cfb"
)

// summaryStream is the OLE stream holding the author properties. The name
// starts with the control character 0x05.
const summaryStream = "\x05SummaryInformation"

// updateLegacy rewrites the author properties of a .doc/.xls/.ppt document.
func updateLegacy(path string, meta Metadata) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	container, err := cfb.Open(data)
	if err != nil {
		return err
	}
	if !container.HasStream(summaryStream) {
		return fmt.Errorf("в документе нет блока свойств SummaryInformation")
	}
	stream, err := container.ReadStream(summaryStream)
	if err != nil {
		return err
	}
	props, err := parsePropertySet(stream)
	if err != nil {
		return err
	}
	if meta.SetAuthor {
		props.setString(pidAuthor, meta.Author)
	}
	if meta.SetLastModifiedBy {
		props.setString(pidLastAuthor, meta.LastModifiedBy)
	}
	if err := container.WriteStream(summaryStream, props.bytes()); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".topor-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(container.Data()); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return replaceFile(tmpName, path)
}
