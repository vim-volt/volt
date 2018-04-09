package cmd

import (
	"errors"
	"flag"
	"os/user"
	"runtime"

	"github.com/vim-volt/volt/config"
	"github.com/vim-volt/volt/logger"
)

var cmdMap = make(map[string]Cmd)

// Cmd represents volt's subcommand interface.
// All subcommands must implement this.
type Cmd interface {
	ProhibitRootExecution(args []string) bool
	Run(args []string) int
	FlagSet() *flag.FlagSet
}

// Run is invoked by main(), each argument means 'volt {subcmd} {args}'.
func Run(subCmd string, args []string) int {
	subCmd, args, err := expandAlias(subCmd, args)
	if err != nil {
		logger.Error(err.Error())
	}
	self, exists := cmdMap[subCmd]
	if !exists {
		logger.Error("Unknown command '" + subCmd + "'")
		return 3
	}
	if self.ProhibitRootExecution(args) {
		err := detectPriviledgedUser()
		if err != nil {
			logger.Error(err.Error())
			return 4
		}
	}
	return self.Run(args)
}

func expandAlias(subCmd string, args []string) (string, []string, error) {
	cfg, err := config.Read()
	if err != nil {
		return "", nil, errors.New("could not read config.toml: " + err.Error())
	}
	if newArgs, exists := cfg.Alias[subCmd]; exists && len(newArgs) > 0 {
		subCmd = newArgs[0]
		args = append(newArgs[1:], args...)
	}
	return subCmd, args, nil
}

// On Windows, this function always returns nil.
// Because if even administrator user creates a file, the file can be
// overwritten by normal user.
// On Linux, if current user's uid == 0, returns non-nil error.
func detectPriviledgedUser() error {
	if runtime.GOOS == "windows" {
		return nil
	}
	u, err := user.Current()
	if err != nil {
		return errors.New("Cannot get current user: " + err.Error())
	}
	if u.Uid == "0" {
		return errors.New(
			"Cannot run this sub command with root priviledge. " +
				"Please run as normal user")
	}
	return nil
}
