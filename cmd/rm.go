package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/vim-volt/go-volt/lockjson"
	"github.com/vim-volt/go-volt/pathutil"
	"github.com/vim-volt/go-volt/transaction"
)

type rmCmd struct{}

func Rm(args []string) int {
	cmd := rmCmd{}

	reposPath, err := cmd.parseArgs(args)
	if err != nil {
		fmt.Println(err.Error())
		return 10
	}

	err = cmd.removeRepos(reposPath)
	if err != nil {
		fmt.Println("Failed to remove repository: " + err.Error())
		return 11
	}

	return 0
}

func (rmCmd) parseArgs(args []string) (string, error) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Println(`
Usage
  volt rm [-help] {repository}

Description
  Uninstall vim plugin and system plugconf files

Options`)
		fs.PrintDefaults()
		fmt.Println()
	}
	fs.Parse(args)

	if len(fs.Args()) == 0 {
		fs.Usage()
		return "", errors.New("repository was not given")
	}

	reposPath, err := pathutil.NormalizeRepository(fs.Args()[0])
	if err != nil {
		return "", err
	}
	return reposPath, nil
}

func (cmd rmCmd) removeRepos(reposPath string) error {
	path := pathutil.FullReposPathOf(reposPath)

	// Remove system plugconf files
	for _, ext := range []string{".vim", ".json"} {
		fn := reposPath + ext
		plugConf := pathutil.SystemPlugConfOf(fn)
		fmt.Println("[INFO] Removing plugconf " + fn + " ...")
		if _, err := os.Stat(plugConf); !os.IsNotExist(err) {
			err = os.Remove(plugConf)
			if err != nil {
				return err
			}
		}
		dir, _ := filepath.Split(plugConf)
		cmd.removeDirs(dir)
	}

	// Remove existing repository
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		fmt.Println("[INFO] Removing " + path + " ...")
		err = os.RemoveAll(path)
		if err != nil {
			return err
		}
		dir, _ := filepath.Split(path)
		cmd.removeDirs(dir)
	} else {
		return errors.New("no repository was installed: " + path)
	}

	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		return err
	}

	// Begin transaction
	err = transaction.Create()
	if err != nil {
		return err
	}
	defer transaction.Remove()
	lockJSON.TrxID++

	// Rewrite lock.json
	newRepos := make([]lockjson.Repos, 0, len(lockJSON.Repos))
	for _, repos := range lockJSON.Repos {
		if repos.Path != reposPath {
			newRepos = append(newRepos, repos)
		}
	}
	lockJSON.Repos = newRepos

	err = lockjson.Write(lockJSON)
	if err != nil {
		return err
	}

	return nil
}

func (cmd rmCmd) removeDirs(dir string) error {
	if err := os.Remove(dir); err != nil {
		return err
	} else {
		parent, _ := filepath.Split(dir)
		return cmd.removeDirs(parent)
	}
}
