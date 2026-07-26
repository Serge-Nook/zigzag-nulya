//go:build !windows

package filedates

import (
	"errors"
	"time"
)

// CreationTimeSupported reports whether the creation date can be changed.
const CreationTimeSupported = false

// ErrCreationUnsupported is returned on systems where the file birth time
// cannot be modified.
var ErrCreationUnsupported = errors.New("изменение даты создания поддерживается только в Windows")

func setCreationTime(string, time.Time) error {
	return ErrCreationUnsupported
}
