package cmd

import (
	"flag"
	"fmt"
	"os"
)

type listFlagsType struct {
	helped bool
}

var listFlags listFlagsType

func init() {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Print(`
Usage
  volt list

Quick example
  $ volt list # will list installed plugins

Description
  This is shortcut of:
  volt profile show {current profile}` + "\n\n")
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
	err := profCmd.doShow(append(
		[]string{"-current"},
	))
	if err != nil {
		return 10
	}

	return 0
}
