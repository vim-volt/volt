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
	fmt.Print(
		" .----------------.  .----------------.  .----------------.  .----------------.\n" +
			"| .--------------. || .--------------. || .--------------. || .--------------. |\n" +
			"| | ____   ____  | || |     ____     | || |   _____      | || |  _________   | |\n" +
			"| ||_  _| |_  _| | || |   .'    `.   | || |  |_   _|     | || | |  _   _  |  | |\n" +
			"| |  \\ \\   / /   | || |  /  .--.  \\  | || |    | |       | || | |_/ | | \\_|  | |\n" +
			"| |   \\ \\ / /    | || |  | |    | |  | || |    | |   _   | || |     | |      | |\n" +
			"| |    \\ ' /     | || |  \\  `--'  /  | || |   _| |__/ |  | || |    _| |_     | |\n" +
			"| |     \\_/      | || |   `.____.'   | || |  |________|  | || |   |_____|    | |\n" +
			"| |              | || |              | || |              | || |              | |\n" +
			"| '--------------' || '--------------' || '--------------' || '--------------' |\n" +
			" '----------------'  '----------------'  '----------------'  '----------------'\n" +
			`
Usage
  volt COMMAND ARGS

Command
  get [-l] [-u] [-v] [{repository} ...]
    Install / Upgrade vim plugin, and fetch plugconf from
    https://github.com/vim-volt/plugconf-templates

  rm {repository} [{repository2} ...]
    Uninstall vim plugin and plugconf files

  add {from} {repository}
    Add local {from} repository as {repository} to lock.json

  enable {repository} [{repository2} ...]
    This is shortcut of:
    volt profile add {current profile} {repository} [{repository2} ...]

  list
    This is shortcut of:
    volt profile show {current profile}

  disable {repository} [{repository2} ...]
    This is shortcut of:
    volt profile rm {current profile} {repository} [{repository2} ...]

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

  profile rename {old} {new}
    Rename profile {old} to {new}

  profile add {name} {repository} [{repository2} ...]
    Add one or more repositories to profile

  profile rm {name} {repository} [{repository2} ...]
    Remove one or more repositories to profile

  profile use [-current | {name}] vimrc [true | false]
  profile use [-current | {name}] gvimrc [true | false]
    Set vimrc / gvimrc flag to true or false.

  build [-full]
    Build ~/.vim/pack/volt/ directory

  migrate
    Convert old version $VOLTPATH/lock.json structure into the latest version

  self-upgrade [-check]
    Upgrade to the latest volt command, or if -check was given, it only checks the newer version is available

  version
    Show volt command version` + "\n\n")
}
