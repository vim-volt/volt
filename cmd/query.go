package cmd

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/pathutil"
)

type queryFlagsType struct {
	helped   bool
	json     bool
	lockJSON bool
}

var queryFlags queryFlagsType

func init() {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Print(`
Usage
  volt query [-help] [-j] [-l] [{repository} ...]

Quick example
  $ volt query -l # show all installed vim plugins info
  $ volt query -l -j # show all installed vim plugins info as JSON
  $ volt query tyru/caw.vim # show tyru/caw.vim plugin info (if tyru/caw.vim is not installed, fetch vim plugin info from remote)

Description
  Output queried vim plugins info ("version" and "trx_id"). "version" is vim plugin's locked version. "trx_id" is transaction ID (transaction is volt's internal operation unit). The ID is incremented when plugins are installed or uninstalled by "volt add", "volt get", "volt rm".

  {repository} is treated as same format as "volt get" (see "volt get -help").` + "\n\n")
		fmt.Println("Options")
		fs.PrintDefaults()
		fmt.Println()
		queryFlags.helped = true
	}
	fs.BoolVar(&queryFlags.json, "j", false, "output as JSON")
	fs.BoolVar(&queryFlags.lockJSON, "l", false, "show installed plugins")

	cmdFlagSet["query"] = fs
}

type queryCmd struct{}

func Query(args []string) int {
	cmd := queryCmd{}

	args, flags, err := cmd.parseArgs(args)
	if err == ErrShowedHelp {
		return 0
	}
	if err != nil {
		logger.Error(err.Error())
		return 10
	}

	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		logger.Error("Failed to read lock.json: " + err.Error())
		return 11
	}

	reposPathList, err := cmd.getReposPathList(flags, args, lockJSON)
	if err != nil {
		logger.Error(err.Error())
		return 12
	}

	reposList := make([]lockjson.Repos, 0, len(reposPathList))
	for _, reposPath := range reposPathList {
		repos, err := lockJSON.Repos.FindByPath(reposPath)
		if err != nil {
			// TODO: show plugin info on remote
			logger.Error("Not implemented yet: remote query")
			return 13
		}
		reposList = append(reposList, *repos)
	}

	cmd.printReposList(reposList, flags)

	return 0
}

func (*queryCmd) parseArgs(args []string) ([]string, *queryFlagsType, error) {
	fs := cmdFlagSet["query"]
	fs.Parse(args)
	if queryFlags.helped {
		return nil, nil, ErrShowedHelp
	}

	if !queryFlags.lockJSON && len(fs.Args()) == 0 {
		fs.Usage()
		return nil, nil, errors.New("repository was not given")
	}

	return fs.Args(), &queryFlags, nil
}

func (*queryCmd) getReposPathList(flags *queryFlagsType, args []string, lockJSON *lockjson.LockJSON) ([]string, error) {
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

func (*queryCmd) printReposList(reposList []lockjson.Repos, flags *queryFlagsType) error {
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
