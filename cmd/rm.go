package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/vim-volt/go-volt/lockjson"
	"github.com/vim-volt/go-volt/pathutil"
)

type rmCmd struct{}

type rmFlags struct {
	removePlugConf bool
}

func Rm(args []string) int {
	cmd := rmCmd{}

	reposPath, flags, err := cmd.parseArgs(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 10
	}

	err = cmd.removeRepos(reposPath, flags)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to clone repository: "+err.Error())
		return 11
	}

	return 0
}

func (rmCmd) parseArgs(args []string) (string, *rmFlags, error) {
	var flags rmFlags
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `
Usage
  volt rm [-help] [-p] {repository}

Description
  Uninstall vim plugin (and plugconf file also if -p was given

Options`)
		fs.PrintDefaults()
		fmt.Fprintln(os.Stderr)
	}
	fs.BoolVar(&flags.removePlugConf, "p", false, "Remove plugconf")
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

func (rmCmd) removeRepos(reposPath string, flags *rmFlags) error {
	path := pathutil.FullReposPathOf(reposPath)

	// Remove plugconf file
	if flags.removePlugConf {
		plugConf := pathutil.PlugConfOf(reposPath)
		fmt.Println("[INFO] Removing plugconf " + plugConf + " ...")
		if _, err := os.Stat(plugConf); !os.IsNotExist(err) {
			err = os.Remove(plugConf)
			if err != nil {
				return err
			}
		}
	}

	// Remove existing repository
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		fmt.Println("[INFO] Removing " + path + " ...")
		err = os.RemoveAll(path)
		if err != nil {
			return err
		}
	} else {
		return errors.New("no repository was installed: " + path)
	}

	// Read lock.json
	json, err := lockjson.Read()
	if err != nil {
		return err
	}

	// Rewrite lock.json
	newRepos := make([]lockjson.Repos, 0, len(json.Repos))
	for _, repos := range json.Repos {
		if repos.Path != reposPath {
			newRepos = append(newRepos, repos)
		}
	}
	json.Repos = newRepos

	err = lockjson.Write(json)
	if err != nil {
		return err
	}

	return nil
}
