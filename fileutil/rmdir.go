package fileutil

import (
	"os"
	"path/filepath"
	"strings"
)

// Always returns non-nil error which is the last error of os.Remove(dir)
func RemoveDirs(dir string) error {
	// Remove trailing slashes
	dir = strings.TrimRight(dir, "/")

	if err := os.Remove(dir); err != nil {
		return err
	} else {
		return RemoveDirs(filepath.Dir(dir))
	}
}
