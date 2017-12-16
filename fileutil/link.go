package fileutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// TryLinkDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
func TryLinkDir(src, dst string, buf []byte, perm os.FileMode, invalidType os.FileMode) error {
	if err := os.MkdirAll(dst, perm); err != nil {
		return err
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

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
			if err = TryLinkDir(srcPath, dstPath, buf, entries[i].Mode(), invalidType); err != nil {
				return err
			}
		} else {
			if err = TryLinkFile(srcPath, dstPath, buf, entries[i].Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}

// TryLinkFile tries os.Link() at first, but if it failed call CopyFile to copy
// the contents of src to dst
func TryLinkFile(src, dst string, buf []byte, perm os.FileMode) error {
	if err := os.Link(src, dst); err == nil {
		return err
	}
	return CopyFile(src, dst, buf, perm)
}
