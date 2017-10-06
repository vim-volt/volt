package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/pathutil"
)

type disableFlagsType struct {
	helped bool
}

var disableFlags disableFlagsType

func init() {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Println(`
Usage
  volt disable {repository} [{repository2} ...]

Quick example
  $ volt disable tyru/caw.vim # will disable tyru/caw.vim plugin in current profile

Description
  This is shortcut of:
  volt profile rm {current profile} {repository} [{repository2} ...]
`)
		//fmt.Println("Options")
		//fs.PrintDefaults()
		fmt.Println()
		disableFlags.helped = true
	}

	cmdFlagSet["disable"] = fs
}

type disableCmd struct{}

func Disable(args []string) int {
	cmd := disableCmd{}

	reposPathList, err := cmd.parseArgs(args)
	if err == ErrShowedHelp {
		return 0
	}
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

	err = profCmd.doRm(append(
		[]string{currentProfile},
		reposPathList...,
	))
	if err != nil {
		return 12
	}

	return 0
}

func (*disableCmd) parseArgs(args []string) ([]string, error) {
	fs := cmdFlagSet["disable"]
	fs.Parse(args)
	if disableFlags.helped {
		return nil, ErrShowedHelp
	}

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
