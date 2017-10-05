package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/pathutil"
)

type enableCmd struct{}

func Enable(args []string) int {
	cmd := enableCmd{}

	reposPathList, err := cmd.parseArgs(args)
	if err != nil {
		logger.Error("Failed to parse args: " + err.Error())
		return 10
	}

	profCmd := profileCmd{}
	currentProfile, err := profCmd.getCurrentProfile()
	if err != nil {
		logger.Error(err.Error())
		return 11
	}

	err = profCmd.doAdd(append(
		[]string{currentProfile},
		reposPathList...,
	))
	if err != nil {
		return 12
	}

	return 0
}

func (*enableCmd) parseArgs(args []string) ([]string, error) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Println(`
Usage
  volt enable {repository} [{repository2} ...]

Quick example
  $ volt enable tyru/caw.vim # will enable tyru/caw.vim plugin in current profile

Description
  This is shortcut of:
  volt profile add {current profile} {repository} [{repository2} ...]

Options`)
		fs.PrintDefaults()
		fmt.Println()
	}
	fs.Parse(args)

	if len(fs.Args()) == 0 {
		fs.Usage()
		return nil, errors.New("repository was not given")
	}

	// Normalize repos path
	var reposPathList []string
	for _, arg := range fs.Args() {
		reposPath, err := pathutil.NormalizeRepos(arg)
		if err != nil {
			return nil, err
		}
		reposPathList = append(reposPathList, reposPath)
	}

	return reposPathList, nil
}
