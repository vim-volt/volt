// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package fileutil

import (
	"fmt"
	"os"
	"syscall"
)

// CopyFile copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file. The file mode is set to perm and
// the copied data is synced/flushed to stable storage.
func CopyFile(src, dst string, buf []byte, perm os.FileMode) error {
	r, err := os.Open(src)
	if err != nil {
		return err
	}
	fi, err := r.Stat()
	if err != nil {
		return err
	}
	w, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}

	wfd := int(w.Fd())
	rfd := int(r.Fd())
	if int64(int(fi.Size())) < fi.Size() {
		var written int64 = 0
		readsize := int(fi.Size())
		for {
			if n, err := syscall.Sendfile(wfd, rfd, nil, readsize); err != nil {
				return fmt.Errorf("sendfile(%q, %q) failed: %s", src, dst, err.Error())
			} else {
				written += int64(n)
				if written >= fi.Size() {
					break
				}
			}
		}
	} else {
		if _, err := syscall.Sendfile(wfd, rfd, nil, int(fi.Size())); err != nil {
			return fmt.Errorf("sendfile(%q, %q) failed: %s", src, dst, err.Error())
		}
	}
	return nil
}
