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
	"io/ioutil"
	"os"
	"path/filepath"
)

// TODO: Parallelism

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
func CopyDir(src, dst string, buf []byte, perm os.FileMode, ignoreType os.FileMode) error {
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
		if entries[i].Mode()&ignoreType != 0 {
			continue
		}

		srcPath := filepath.Join(src, entries[i].Name())
		dstPath := filepath.Join(dst, entries[i].Name())

		if entries[i].IsDir() {
			if err = CopyDir(srcPath, dstPath, buf, entries[i].Mode(), ignoreType); err != nil {
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
