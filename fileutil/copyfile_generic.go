// +build !darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris

package fileutil

import (
	"io"
	"os"
)

// CopyFile copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file. The file mode is set to perm and
// the copied data is synced/flushed to stable storage.
func CopyFile(src, dst string, buf []byte, perm os.FileMode) (err error) {
	r, err := os.Open(src)
	if err != nil {
		return
	}
	defer func() {
		if e := r.Close(); e != nil {
			err = e
		}
	}()

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
