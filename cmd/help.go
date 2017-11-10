package cmd

import (
	"errors"
	"flag"
	"fmt"

	"github.com/vim-volt/volt/logger"
)

var cmdFlagSet = make(map[string]*flag.FlagSet)

var ErrShowedHelp = errors.New("already showed help")

func Help(args []string) int {
	if len(args) == 0 {
		showHelp()
		return 0
	}
	if args[0] == "help" { // "volt help help"
		fmt.Println("E478: Don't panic!")
		return 0
	}

	if fs, exists := cmdFlagSet[args[0]]; exists {
		fs.Usage()
		return 0
	} else {
		logger.Errorf("Unknown command '%s'", args[0])
		return 1
	}
}

func showHelp() {
	fmt.Print(`
Usage
  volt COMMAND ARGS

Command
  get [-l] [-u] [-v] [{repository} ...]
    Install / Upgrade vim plugin, and system plugconf files from
    https://github.com/vim-volt/plugconf-templates

  rm {repository} [{repository2} ...]
    Uninstall vim plugin and system plugconf files

  add {from} {repository}
    Add local {from} repository as {repository} to lock.json

  query [-j] [-l] [{repository} ...]
    Output queried vim plugin info

  enable {repository} [{repository2} ...]
    This is shortcut of:
    volt profile add {current profile} {repository} [{repository2} ...]

  list
    This is shortcut of:
    volt profile show {current profile}

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

  profile use [-current | {name}] vimrc [true | false]
  profile use [-current | {name}] gvimrc [true | false]
    Set vimrc / gvimrc flag to true or false.

  plugconf list [-a]
    List all user plugconfs. If -a option was given, list also system plugconfs.

  plugconf bundle
    Outputs bundled plugconf to stdout.

  plugconf unbundle
    Input bundled plugconf (volt plugconf bundle) from stdin, unbundle the plugconf, and put files to each plugin's plugconf.

  rebuild [-full]
    Rebuild ~/.vim/pack/volt/ directory

  version
    Show volt command version` + "\n\n")
}
