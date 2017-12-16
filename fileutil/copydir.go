package fileutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// TODO: Parallelism

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
func CopyDir(src, dst string, buf []byte, perm os.FileMode, invalidType os.FileMode) error {
	if err := os.MkdirAll(dst, perm); err != nil {
		return err
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	// Avoid allocating in io.Copy() in CopyFile() each time
	if buf == nil {
		buf = make([]byte, 32*1024)
	}

	for i := range entries {
		if entries[i].Mode()&invalidType != 0 {
			return newInvalidType(entries[i].Name())
		}

		srcPath := filepath.Join(src, entries[i].Name())
		dstPath := filepath.Join(dst, entries[i].Name())

		if entries[i].IsDir() {
			if err = CopyDir(srcPath, dstPath, buf, entries[i].Mode(), invalidType); err != nil {
				return err
			}
		} else {
			if err = CopyFile(srcPath, dstPath, buf, entries[i].Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}
