package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/vim-volt/volt/logger"
)

type listFlagsType struct {
	helped bool
}

var listFlags listFlagsType

func init() {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Println(`
Usage
  volt list

Quick example
  $ volt list # will list installed plugins

Description
  This is shortcut of:
  volt profile show {current profile}
`)
		//fmt.Println("Options")
		//fs.PrintDefaults()
		fmt.Println()
		listFlags.helped = true
	}

	cmdFlagSet["list"] = fs
}

type listCmd struct{}

func List(args []string) int {
	profCmd := profileCmd{}
	currentProfile, err := profCmd.getCurrentProfile()
	if err != nil {
		logger.Error(err.Error())
		return 10
	}

	err = profCmd.doShow(append(
		[]string{currentProfile},
	))
	if err != nil {
		return 11
	}

	return 0
}
