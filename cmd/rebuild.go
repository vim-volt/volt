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

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
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

	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		return errors.New("could not read lock.json: " + err.Error())
	}

	reposList, err := cmd.getActiveProfileRepos(lockJSON)
	if err != nil {
		return err
	}

	fmt.Println("[INFO] Rebuilding ~/.vim/pack/volt directory ...")

	var removeDone <-chan error
	if _, err := os.Stat(startDir); !os.IsNotExist(err) {
		var err error
		removeDone, err = cmd.removeStartDir(startDir)
		if err != nil {
			return err
		}
	}

	err = os.MkdirAll(startDir, 0755)
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
	if removeDone != nil {
		err = <-removeDone
		if err != nil {
			return errors.New("failed to remove '" + startDir + "': " + err.Error())
		}
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
	profile, err := lockJSON.Profiles.FindByName(lockJSON.ActiveProfile)
	if err != nil {
		// this must not be occurred because lockjson.Read()
		// validates that the matching profile exists
		return nil, err
	}

	return lockJSON.GetReposListByProfile(profile)
}

func (cmd *rebuildCmd) copyRepos(repos *lockjson.Repos, startDir string, done chan copyReposResult) {
	src := pathutil.FullReposPathOf(repos.Path)
	dst := filepath.Join(startDir, cmd.encodeReposPath(repos.Path))

	r, err := git.PlainOpen(src)
	if err != nil {
		done <- copyReposResult{
			errors.New("failed to open repository: " + err.Error()),
			repos,
		}
		return
	}

	head, err := r.Head()
	if err != nil {
		done <- copyReposResult{
			errors.New("failed to get HEAD reference: " + err.Error()),
			repos,
		}
		return
	}

	commit, err := r.CommitObject(head.Hash())
	if err != nil {
		done <- copyReposResult{
			errors.New("failed to get HEAD commit object: " + err.Error()),
			repos,
		}
		return
	}

	tree, err := r.TreeObject(commit.TreeHash)
	if err != nil {
		done <- copyReposResult{
			errors.New("failed to get tree " + head.Hash().String() + ": " + err.Error()),
			repos,
		}
		return
	}

	err = tree.Files().ForEach(func(file *object.File) error {
		osMode, err := file.Mode.ToOSFileMode()
		if err != nil {
			return errors.New("failed to convert file mode: " + err.Error())
		}

		contents, err := file.Contents()
		if err != nil {
			return errors.New("failed get file contents: " + err.Error())
		}

		filename := filepath.Join(dst, file.Name)
		dir, _ := filepath.Split(filename)
		os.MkdirAll(dir, 0755)
		ioutil.WriteFile(filename, []byte(contents), osMode)
		return nil
	})
	if err != nil {
		done <- copyReposResult{err, repos}
		return
	}

	fmt.Println("[INFO] Copying repository " + repos.Path + " ... Done.")

	done <- copyReposResult{nil, repos}
}

func (*rebuildCmd) encodeReposPath(reposPath string) string {
	return strings.NewReplacer("_", "__", "/", "_").Replace(reposPath)
}
