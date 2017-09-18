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
	if len(os.Args) <= 1 {
		showUsage()
		return 1
	}
	switch os.Args[1] {
	case "get":
		return cmd.Get(os.Args[2:])
	case "rm":
		return cmd.Rm(os.Args[2:])
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
  config [-global] {key} [{value}]
    Set / Get config value

  get [-u] [-v] {repository}
    Install / Upgrade vim plugin

  rm [-p] {repository}
    Uninstall vim plugin

  query [-json] [-installed] {repository}
    Output queried vim plugin info

  plugconf ping {repository}
    Check if plugconf file of {repository} exists on
    https://github.com/vim-volt/plugconf-templates

  plugconf get {repository}
    Install recommended plugconf file of {repository} from
    https://github.com/vim-volt/plugconf-templates
`)
}
