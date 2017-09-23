package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/vim-volt/go-volt/lockjson"
	"github.com/vim-volt/go-volt/pathutil"
	"github.com/vim-volt/go-volt/transaction"
)

type rmCmd struct{}

type rmFlags struct {
	removePlugConf bool
}

func Rm(args []string) int {
	cmd := rmCmd{}

	reposPath, flags, err := cmd.parseArgs(args)
	if err != nil {
		fmt.Println(err.Error())
		return 10
	}

	err = cmd.removeRepos(reposPath, flags)
	if err != nil {
		fmt.Println("Failed to remove repository: " + err.Error())
		return 11
	}

	return 0
}

func (rmCmd) parseArgs(args []string) (string, *rmFlags, error) {
	var flags rmFlags
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Println(`
Usage
  volt rm [-help] [-p] {repository}

Description
  Uninstall vim plugin (and plugconf file also if -p was given

Options`)
		fs.PrintDefaults()
		fmt.Println()
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
