package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	// Remove system plugconf (vim, json)
	fmt.Println("[INFO] Removing plugconf files ...")
	err = cmd.removeSystemPlugConf(reposPath)
	if err != nil {
		return err
	}

	// Remove parent directories of system plugconf
	dir, _ := filepath.Split(pathutil.SystemPlugConfOf(reposPath))
	err = cmd.removeDirs(dir)

	// Remove existing repository
	fullpath := pathutil.FullReposPathOf(reposPath)
	if _, err = os.Stat(fullpath); !os.IsNotExist(err) {
		fmt.Println("[INFO] Removing " + fullpath + " ...")
		err = os.RemoveAll(fullpath)
		if err != nil {
			return err
		}
		dir, _ := filepath.Split(fullpath)
		cmd.removeDirs(dir)
	} else {
		return errors.New("no repository was installed: " + fullpath)
	}

	// Delete repos path from lockJSON.Repos[i]
	for i := range lockJSON.Repos {
		if lockJSON.Repos[i].Path == reposPath {
			lockJSON.Repos = append(lockJSON.Repos[:i], lockJSON.Repos[i+1:]...)
			break
		}
	}

	// Delete repos path from profiles[i]/repos_path[j]
	for i, profile := range lockJSON.Profiles {
		for j, profReposPath := range profile.ReposPath {
			if profReposPath == reposPath {
				lockJSON.Profiles[i].ReposPath = append(
					lockJSON.Profiles[i].ReposPath[:j],
					lockJSON.Profiles[i].ReposPath[j+1:]...,
				)
				break
			}
		}
	}

	// Write to lock.json
	err = lockjson.Write(lockJSON)
	if err != nil {
		return err
	}

	return nil
}

func (cmd rmCmd) removeSystemPlugConf(reposPath string) error {
	for _, ext := range []string{".vim", ".json"} {
		plugConf := pathutil.SystemPlugConfOf(reposPath + ext)
		if _, err := os.Stat(plugConf); !os.IsNotExist(err) {
			err = os.Remove(plugConf)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (cmd rmCmd) removeDirs(dir string) error {
	// Remove trailing slashes
	dir = strings.TrimRight(dir, "/")

	if err := os.Remove(dir); err != nil {
		return err
	} else {
		parent, _ := filepath.Split(dir)
		return cmd.removeDirs(parent)
	}
}
