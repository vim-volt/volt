package builder

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/vim-volt/volt/cmd/buildinfo"
	"github.com/vim-volt/volt/fileutil"
	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/pathutil"
)

type BaseBuilder struct{}

func (builder *BaseBuilder) installVimrcAndGvimrc(profileName, vimrcPath, gvimrcPath string) error {
	// Save old vimrc file as {vimrc}.bak
	vimrcInfo, err := os.Stat(vimrcPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	vimrcExists := !os.IsNotExist(err)
	if vimrcExists {
		err = fileutil.CopyFile(vimrcPath, vimrcPath+".bak", make([]byte, vimrcInfo.Size()), vimrcInfo.Mode())
		if err != nil {
			return err
		}
	}
	defer os.Remove(vimrcPath + ".bak")

	// Install vimrc
	err = builder.installRCFile(
		profileName,
		pathutil.ProfileVimrc,
		vimrcPath,
	)
	if err != nil {
		return err
	}

	// Install gvimrc
	err = builder.installRCFile(
		profileName,
		pathutil.ProfileGvimrc,
		gvimrcPath,
	)
	if err != nil {
		// Restore old vimrc
		if vimrcExists {
			err2 := os.Rename(vimrcPath+".bak", vimrcPath)
			if err2 != nil {
				return multierror.Append(err, err2)
			}
		} else {
			err2 := os.Remove(vimrcPath)
			if err2 != nil {
				return multierror.Append(err, err2)
			}
		}
		return err
	}
	return nil
}

func (builder *BaseBuilder) installRCFile(profileName, srcRCFileName, dst string) error {
	src := filepath.Join(pathutil.RCDir(profileName), srcRCFileName)

	// Return error if destination file does not have magic comment
	if pathutil.Exists(dst) {
		// If the file does not have magic comment
		if !builder.HasMagicComment(dst) {
			if !pathutil.Exists(src) {
				return nil
			}
			return fmt.Errorf("'%s' is not an auto-generated file. please move to '%s' and re-run 'volt build'", dst, pathutil.RCDir(profileName))
		}
	}

	// Remove destination (~/.vim/vimrc or ~/.vim/gvimrc)
	os.Remove(dst)
	if pathutil.Exists(dst) {
		return errors.New("failed to remove " + dst)
	}

	// Skip if rc file does not exist
	if !pathutil.Exists(src) {
		return nil
	}

	return builder.copyFileWithMagicComment(src, dst)
}

const magicComment = "\" NOTE: this file was generated by volt. please modify original file.\n"
const magicCommentNext = "\" Original file: %s\n\n"

// Return error if the magic comment does not exist
func (*BaseBuilder) HasMagicComment(dst string) bool {
	r, err := os.Open(dst)
	if err != nil {
		return false
	}
	defer r.Close()

	magic := []byte(magicComment)
	read := make([]byte, len(magic))
	n, err := r.Read(read)
	if err != nil || n < len(magicComment) {
		return false
	}

	for i := range magic {
		if magic[i] != read[i] {
			return false
		}
	}
	return true
}

func (builder *BaseBuilder) copyFileWithMagicComment(src, dst string) (err error) {
	r, err := os.Open(src)
	if err != nil {
		return
	}
	defer func() {
		if e := r.Close(); e != nil {
			err = e
		}
	}()

	os.MkdirAll(filepath.Dir(dst), 0755)
	w, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := w.Close(); e != nil {
			err = e
		}
	}()

	_, err = w.Write([]byte(magicComment))
	if err != nil {
		return
	}
	_, err = w.Write([]byte(fmt.Sprintf(magicCommentNext, src)))
	if err != nil {
		return
	}

	_, err = io.Copy(w, r)
	return
}

type actionReposResult struct {
	err   error
	repos *lockjson.Repos
	files buildinfo.FileMap
}

func (builder *BaseBuilder) helptags(reposPath pathutil.ReposPath, vimExePath string) error {
	// Do nothing if <reposPath>/doc directory doesn't exist
	docdir := filepath.Join(reposPath.EncodeToPlugDirName(), "doc")
	if !pathutil.Exists(docdir) {
		return nil
	}
	// Execute ":helptags doc" in reposPath
	vimArgs := builder.makeVimArgs(reposPath)
	logger.Debugf("Executing '%s %s' ...", vimExePath, strings.Join(vimArgs, " "))
	err := exec.Command(vimExePath, vimArgs...).Run()
	if err != nil {
		return errors.New("failed to make tags file: " + err.Error())
	}
	return nil
}

func (*BaseBuilder) makeVimArgs(reposPath pathutil.ReposPath) []string {
	path := reposPath.EncodeToPlugDirName()
	return []string{
		"-u", "NONE", "-i", "NONE", "-N",
		"--cmd", "cd " + path,
		"--cmd", "set rtp+=" + path,
		"--cmd", "helptags doc",
		"--cmd", "quit",
	}
}
