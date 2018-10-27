// +build go1.9

package main

import (
	"os"
	"os/user"
	"runtime"

	"github.com/pkg/errors"
	"github.com/vim-volt/volt/config"
	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/subcmd"
)

func main() {
	code, msg := run(os.Args)
	if code != 0 {
		logger.Error(msg)
		os.Exit(code)
	}
}

func run(args []string) (int, string) {
	if os.Getenv("VOLT_DEBUG") != "" {
		logger.SetLevel(logger.DebugLevel)
	}

	if len(args) <= 1 {
		args = append(args, "help")
	}
	subCmd := args[1]
	args = args[2:]

	// Expand subcommand alias
	subCmd, args, err := expandAlias(subCmd, args)
	if err != nil {
		return 1, err.Error()
	}

	c := subcmd.LookUpCmd(subCmd)
	if c == nil {
		return 3, "Unknown command '" + subCmd + "'"
	}

	// Disallow executing the commands which may modify files in root priviledge
	if c.ProhibitRootExecution(args) {
		err := detectPriviledgedUser()
		if err != nil {
			return 4, err.Error()
		}
	}

	lockJSON, err := lockjson.Read()
	if err != nil {
		return 20, errors.Wrap(err, "failed to read lock.json").Error()
	}

	cfg, err := config.Read()
	if err != nil {
		return 30, errors.Wrap(err, "failed to read config.toml").Error()
	}

	result := c.Run(&subcmd.CmdContext{
		Args:     args,
		LockJSON: lockJSON,
		Config:   cfg,
	})
	if result != nil {
		return result.Code, result.Msg
	}
	return 0, ""
}

func expandAlias(subCmd string, args []string) (string, []string, error) {
	cfg, err := config.Read()
	if err != nil {
		return "", nil, errors.Wrap(err, "could not read config.toml")
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
		return errors.Wrap(err, "cannot get current user")
	}
	if u.Uid == "0" {
		return errors.New(
			"cannot run this sub command with root priviledge. " +
				"Please run as normal user")
	}
	return nil
}
