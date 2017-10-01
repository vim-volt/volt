package main

import (
	"fmt"
	"os"

	"github.com/vim-volt/volt/cmd"
	"github.com/vim-volt/volt/logger"
)

func main() {
	os.Exit(Main())
}

func Main() int {
	if len(os.Args) <= 1 ||
		(len(os.Args) == 1 && os.Args[1] == "help") {
		showUsage()
		return 1
	}
	switch os.Args[1] {
	case "get":
		return cmd.Get(os.Args[2:])
	case "rm":
		return cmd.Rm(os.Args[2:])
	case "add":
		return cmd.Add(os.Args[2:])
	case "query":
		return cmd.Query(os.Args[2:])
	case "enable":
		return cmd.Enable(os.Args[2:])
	case "disable":
		return cmd.Disable(os.Args[2:])
	case "profile":
		return cmd.Profile(os.Args[2:])
	case "rebuild":
		return cmd.Rebuild(os.Args[2:])
	case "version":
		return cmd.Version(os.Args[2:])
	case "help":
		logger.Errorf("Run 'volt %s -help' to see its help", os.Args[2])
		return 2
	default:
		logger.Error("Unknown command '" + os.Args[1] + "'")
		return 3
	}
}

func showUsage() {
	fmt.Println(`
Usage
  volt COMMAND ARGS

Command
  get [-l] [-u] [-v] {repository}
    Install / Upgrade vim plugin, and system plugconf files from
    https://github.com/vim-volt/plugconf-templates

  rm {repository}
    Uninstall vim plugin and system plugconf files

  add {repository}
    Add local {repository} to lock.json

  add {from} {repository}
    Add local {from} repository as {repository} to lock.json

  query [-j] [-l] [{repository}]
    Output queried vim plugin info

  enable {repository} [{repository2} ...]
    This is shortcut of:
    volt profile add {current profile} {repository} [{repository2} ...]

  disable {repository} [{repository2} ...]
    This is shortcut of:
    volt profile rm {current profile} {repository} [{repository2} ...]

  profile [get]
    Get current profile name

  profile set {name}
    Set profile name

  profile show {name}
    Show profile info

  profile list
    List all profiles

  profile new {name}
    Create new profile

  profile destroy {name}
    Delete profile

  profile add {name} {repository} [{repository2} ...]
    Add one or more repositories to profile

  profile rm {name} {repository} [{repository2} ...]
    Remove one or more repositories to profile

  rebuild [-full]
    Rebuild ~/.vim/pack/volt/ directory

  version
    Show volt command version
`)
}
