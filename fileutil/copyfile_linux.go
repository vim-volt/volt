// +build linux

/* MIT License
 *
 * Copyright (c) 2017 Roland Singer [roland.singer@desertbit.com]
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

/*
 * See the following URL for the original code:
 *   https://gist.github.com/r0l1/92462b38df26839a3ca324697c8cba04
 */

package fileutil

import (
	"os"
	"syscall"

	"github.com/pkg/errors"
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
		written := int64(0)
		readsize := int(fi.Size())
		for {
			n, err := syscall.Sendfile(wfd, rfd, nil, readsize)
			if err != nil {
				return errors.Errorf("sendfile(%q, %q) failed: %s", src, dst, err.Error())
			}
			written += int64(n)
			if written >= fi.Size() {
				break
			}
		}
	} else {
		if _, err := syscall.Sendfile(wfd, rfd, nil, int(fi.Size())); err != nil {
			return errors.Errorf("sendfile(%q, %q) failed: %s", src, dst, err.Error())
		}
	}
	return nil
}
