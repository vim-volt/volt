package cmd

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/vim-volt/go-volt/lockjson"
	"github.com/vim-volt/go-volt/pathutil"
)

type queryCmd struct{}

type queryFlags struct {
	json     bool
	lockJSON bool
}

func Query(args []string) int {
	cmd := queryCmd{}

	args, flags, err := cmd.parseArgs(args)
	if err != nil {
		fmt.Println(err.Error())
		return 10
	}

	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		fmt.Println("[ERROR] Failed to read lock.json: " + err.Error())
		return 11
	}

	reposPathList, err := cmd.getReposPathList(flags, args, lockJSON)
	if err != nil {
		fmt.Println(err.Error())
		return 12
	}

	reposList := make([]lockjson.Repos, 0, len(reposPathList))
	for _, reposPath := range reposPathList {
		repos, err := lockJSON.Repos.FindByPath(reposPath)
		if err != nil {
			// TODO: show plugin info on remote
			fmt.Println("[ERROR] Not implemented yet: remote query")
			return 13
		}
		reposList = append(reposList, *repos)
	}

	cmd.printReposList(reposList, flags)

	return 0
}

func (queryCmd) parseArgs(args []string) ([]string, *queryFlags, error) {
	var flags queryFlags
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Println(`
Usage
  volt query [-help] [-j] [-l] [{repository}]

Description
  Output queried vim plugin info

Options`)
		fs.PrintDefaults()
		fmt.Println()
	}
	fs.BoolVar(&flags.json, "j", false, "output as JSON")
	fs.BoolVar(&flags.lockJSON, "l", false, "show installed plugins")
	fs.Parse(args)

	if !flags.lockJSON && len(fs.Args()) == 0 {
		fs.Usage()
		return nil, nil, errors.New("repository was not given")
	}

	return fs.Args(), &flags, nil
}

func (queryCmd) getReposPathList(flags *queryFlags, args []string, lockJSON *lockjson.LockJSON) ([]string, error) {
	reposPathList := make([]string, 0, 32)
	if flags.lockJSON {
		for _, repos := range lockJSON.Repos {
			reposPathList = append(reposPathList, repos.Path)
		}
	}
	for _, arg := range args {
		reposPath, err := pathutil.NormalizeRepos(arg)
		if err != nil {
			return nil, err
		}
		reposPathList = append(reposPathList, reposPath)
	}
	return reposPathList, nil
}

func (queryCmd) printReposList(reposList []lockjson.Repos, flags *queryFlags) error {
	if flags.json {
		bytes, err := json.Marshal(reposList)
		if err != nil {
			return err
		}
		fmt.Print(string(bytes))
	} else {
		for _, repos := range reposList {
			fmt.Println(repos.Path)
			fmt.Println("  version:", repos.Version)
			fmt.Println("  trx_id:", repos.TrxID)
		}
	}

	return nil
}
