package fileutil

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
// Symlinks are ignored and skipped.
func CopyDir(src, dst string, buf []byte, perm os.FileMode) error {
	if err := os.MkdirAll(dst, perm); err != nil {
		return err
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	for i := range entries {
		srcPath := filepath.Join(src, entries[i].Name())
		dstPath := filepath.Join(dst, entries[i].Name())

		si, err := os.Stat(srcPath)
		if err != nil {
			return err
		}

		if entries[i].IsDir() {
			if err = CopyDir(srcPath, dstPath, buf, si.Mode()); err != nil {
				return err
			}
		} else if entries[i].Mode()&os.ModeSymlink != 0 {
			// Skip symlinks
		} else {
			if err = CopyFile(srcPath, dstPath, buf, si.Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}

// CopyFile copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file. The file mode will be copied from the source and
// the copied data is synced/flushed to stable storage.
func CopyFile(src, dst string, buf []byte, perm os.FileMode) (err error) {
	r, err := os.Open(src)
	if err != nil {
		return
	}
	defer r.Close()

	w, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return
	}
	defer func() {
		if e := w.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.CopyBuffer(w, r, buf)
	if err != nil {
		return
	}

	err = w.Sync()
	return
}

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
