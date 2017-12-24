package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/vim-volt/volt/logger"
)

func init() {
	cmdMap["list"] = &listCmd{}
}

type listCmd struct {
	helped bool
}

func (cmd *listCmd) FlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Print(`
Usage
  volt list [-help]

Quick example
  $ volt list # will list installed plugins

Description
  This is shortcut of:
  volt profile show {current profile}` + "\n\n")
		//fmt.Println("Options")
		//fs.PrintDefaults()
		fmt.Println()
		cmd.helped = true
	}
	return fs
}

func (cmd *listCmd) Run(args []string) int {
	fs := cmd.FlagSet()
	fs.Parse(args)
	if cmd.helped {
		return 0
	}

	profCmd := profileCmd{}
	err := profCmd.doShow(append(
		[]string{"-current"},
	))
	if err != nil {
		logger.Error(err.Error())
		return 10
	}

	return 0
}
