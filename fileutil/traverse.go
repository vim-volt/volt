package fileutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func Traverse(dir string, fn func(os.FileInfo)) error {
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for i := range entries {
		if entries[i].IsDir() {
			filename := filepath.Join(dir, entries[i].Name())
			err = Traverse(filename, fn)
			if err != nil {
				return err
			}
		} else {
			fn(entries[i])
		}
	}
	return nil
}
