package fileutil

import (
	"os"
	"path/filepath"
	"strings"
)

// RemoveDirs removes dir and its parent directories until the directory is not
// empty. This always returns non-nil error which is the last error of
// os.Remove(dir).
func RemoveDirs(dir string) error {
	// Remove trailing slashes
	dir = strings.TrimRight(dir, "/")
	if err := os.Remove(dir); err != nil {
		return err
	}
	return RemoveDirs(filepath.Dir(dir))
}
