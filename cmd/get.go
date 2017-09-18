package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/vim-volt/go-volt/lockjson"
	"github.com/vim-volt/go-volt/pathutil"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/protocol/packp/sideband"
)

type getCmd struct{}

type getFlags struct {
	upgrade bool
	verbose bool
}

func Get(args []string) int {
	cmd := getCmd{}

	reposPath, flags, err := cmd.parseArgs(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 10
	}

	err = cmd.cloneRepos(reposPath, flags)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to clone repository: "+err.Error())
		return 11
	}

	return 0
}

func (getCmd) parseArgs(args []string) (string, *getFlags, error) {
	var flags getFlags
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `
Usage
  volt get [-help] [-u] [-v] {repository}

Description
  Install / Upgrade(-u) vim plugin.

Options`)
		fs.PrintDefaults()
		fmt.Fprintln(os.Stderr)
	}
	fs.BoolVar(&flags.upgrade, "u", false, "upgrade installed vim plugin")
	fs.BoolVar(&flags.verbose, "v", false, "show git-clone output")
	fs.Parse(args)

	if len(fs.Args()) == 0 {
		fs.Usage()
		return "", nil, errors.New("repository was not given")
	}

	reposPath, err := pathutil.NormalizeRepository(fs.Args()[0])
	if err != nil {
		return "", nil, err
	}
	return reposPath, &flags, nil
}

func (getCmd) cloneRepos(reposPath string, flags *getFlags) error {
	// Read lock.json
	json, err := lockjson.Read()
	if err != nil {
		return err
	}

	// Return if the same repos path exists
	for _, repos := range json.Repos {
		if repos.Path == reposPath {
			return errors.New("same repos path exists in lock.json: " + reposPath)
		}
	}

	path := pathutil.FullReposPathOf(reposPath)

	// Remove existing repository if -u was given
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		if flags.upgrade {
			fmt.Println("[INFO] Upgrading " + path + " ...")
			err = os.RemoveAll(path)
			if err != nil {
				return errors.New("Failed to remove " + path + ": " + err.Error())
			}
		} else {
			return errors.New("directory already exists: " + path)
		}
	} else {
		fmt.Println("[INFO] Installing " + path + " ...")
	}

	// Mkdir directories
	err = os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return err
	}

	var progress sideband.Progress = nil
	if flags.verbose {
		progress = os.Stdout
	}

	// git clone
	gitRepos, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:      pathutil.CloneURLOf(reposPath),
		Progress: progress,
	})
	if err != nil {
		return err
	}

	// Get HEAD commit hash
	head, err := gitRepos.Head()
	if err != nil {
		return err
	}

	// Rewrite lock.json
	json.Repos = append(json.Repos, lockjson.Repos{
		Path:    reposPath,
		Version: head.Hash().String(),
		Active:  true,
	})
	err = lockjson.Write(json)
	if err != nil {
		return err
	}

	return nil
}
