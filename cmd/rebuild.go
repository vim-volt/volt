package cmd

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/vim-volt/go-volt/lockjson"
	"github.com/vim-volt/go-volt/pathutil"
	"github.com/vim-volt/go-volt/transaction"
)

type rebuildCmd struct{}

func Rebuild(args []string) int {
	// Begin transaction
	err := transaction.Create()
	if err != nil {
		fmt.Println("[ERROR] Failed to begin transaction:", err.Error())
		return 10
	}
	defer transaction.Remove()

	cmd := rebuildCmd{}
	err = cmd.doRebuild()
	if err != nil {
		fmt.Println("[ERROR] Failed to rebuild:", err.Error())
		return 11
	}

	return 0
}

func (cmd *rebuildCmd) doRebuild() error {
	startDir := filepath.Join(
		pathutil.VimDir(), "pack", "volt", "start",
	)

	var removeDone <-chan error
	if _, err := os.Stat(startDir); !os.IsNotExist(err) {
		var err error
		removeDone, err = cmd.removeStartDir(startDir)
		if err != nil {
			return err
		}
	}

	err := os.MkdirAll(startDir, 0755)
	if err != nil {
		return err
	}

	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		return errors.New("could not read lock.json: " + err.Error())
	}

	reposList, err := cmd.getActiveProfileRepos(lockJSON)
	if err != nil {
		return err
	}

	fmt.Println("[INFO] Copying all repositories files to " + startDir + " ...")

	// Copy all repositories files to startDir
	copyDone := make(chan copyReposResult, len(reposList))
	for i := range reposList {
		go cmd.copyRepos(&reposList[i], startDir, copyDone)
	}

	// Wait remove
	err = <-removeDone
	if err != nil {
		return errors.New("failed to remove '" + startDir + "': " + err.Error())
	}

	// Wait copy
	for i := 0; i < len(reposList); i++ {
		result := <-copyDone
		if result.err != nil {
			return errors.New("failed to copy repository '" + result.repos.Path + "': " + result.err.Error())
		}
	}

	return nil
}

func (*rebuildCmd) removeStartDir(startDir string) (<-chan error, error) {
	// Rename startDir to {startDir}.bak
	err := os.Rename(startDir, startDir+".old")
	if err != nil {
		return nil, err
	}

	fmt.Println("[INFO] Removing " + startDir + " ...")

	// Remove files in parallel
	done := make(chan error, 1)
	go func() {
		err = os.RemoveAll(startDir + ".old")
		done <- err
	}()
	return done, nil
}

type copyReposResult struct {
	err   error
	repos *lockjson.Repos
}

func (*rebuildCmd) getActiveProfileRepos(lockJSON *lockjson.LockJSON) ([]lockjson.Repos, error) {
	// Find active profile
	var profile *lockjson.Profile
	for i, p := range lockJSON.Profiles {
		if p.Name == lockJSON.ActiveProfile {
			profile = &lockJSON.Profiles[i]
			break
		}
	}
	if profile == nil {
		// this must not be occurred because lockjson.Read()
		// validates if the matching profile exists
		return nil, errors.New("active profile '" + lockJSON.ActiveProfile + "' does not exist")
	}

	var reposList []lockjson.Repos
	for _, reposPath := range profile.ReposPath {
		var repos *lockjson.Repos
		for i, r := range lockJSON.Repos {
			if r.Path == reposPath {
				repos = &lockJSON.Repos[i]
				break
			}
		}
		if repos == nil {
			// this must not be occurred because lockjson.Read()
			// validates if the matching repos exists
			return nil, errors.New("repos '" + reposPath + "' does not exist")
		}
		reposList = append(reposList, *repos)
	}
	return reposList, nil
}

func (cmd *rebuildCmd) copyRepos(repos *lockjson.Repos, startDir string, done chan copyReposResult) {
	src := pathutil.FullReposPathOf(repos.Path)
	dst := filepath.Join(startDir, cmd.encodeReposPath(repos.Path))
	err := cmd.copyDir(src, dst)
	done <- copyReposResult{err, repos}
}

func (*rebuildCmd) encodeReposPath(reposPath string) string {
	return strings.NewReplacer("_", "__", "/", "_").Replace(reposPath)
}

// CopyFile copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file. The file mode will be copied from the source and
// the copied data is synced/flushed to stable storage.
func (cmd *rebuildCmd) copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
// Symlinks are ignored and skipped.
func (cmd *rebuildCmd) copyDir(src string, dst string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory: " + src)
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	if err == nil {
		return fmt.Errorf("destination already exists: " + dst)
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = cmd.copyDir(srcPath, dstPath)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = cmd.copyFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}

	return
}
