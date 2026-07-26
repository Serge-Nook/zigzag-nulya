//go:build windows

package filedates

import (
	"time"

	"golang.org/x/sys/windows"
)

// CreationTimeSupported reports whether the creation date can be changed.
const CreationTimeSupported = true

func setCreationTime(path string, t time.Time) error {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	handle, err := windows.CreateFile(p, windows.FILE_WRITE_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(handle)

	ft := windows.NsecToFiletime(t.UnixNano())
	return windows.SetFileTime(handle, &ft, nil, nil)
}
