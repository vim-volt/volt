package main

import (
	"fmt"
	"os"

	"github.com/vim-volt/go-volt/cmd"
)

func main() {
	os.Exit(Main())
}

func Main() int {
	if len(os.Args) <= 1 || os.Args[1] == "help" {
		showUsage()
		return 1
	}
	switch os.Args[1] {
	case "get":
		return cmd.Get(os.Args[2:])
	case "rm":
		return cmd.Rm(os.Args[2:])
	case "query":
		return cmd.Query(os.Args[2:])
	case "profile":
		return cmd.Profile(os.Args[2:])
	case "version":
		return cmd.Version(os.Args[2:])
	default:
		fmt.Fprintln(os.Stderr, "[ERROR] Unknown command '"+os.Args[1]+"'")
		return 2
	}
}

func showUsage() {
	fmt.Println(`
Usage
  volt COMMAND ARGS

Command
  get [-l] [-u] [-v] {repository}
    Install / Upgrade vim plugin

  rm [-p] {repository}
    Uninstall vim plugin

  query [-j] [-l] [{repository}]
    Output queried vim plugin info

  profile [{name}]
    Get / Set profile name

  version
    Show volt command version
`)
}
