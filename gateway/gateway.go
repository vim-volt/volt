package gateway

import (
	"errors"
	"os"
	"os/user"
	"runtime"

	"github.com/vim-volt/volt/config"
	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/subcmd"
)

// RunnerFunc invokes c with args.
// On unit testing, a mock function was given.
type RunnerFunc func(c subcmd.Cmd, runctx *subcmd.RunContext) *subcmd.Error

// DefaultRunner simply runs command with args
func DefaultRunner(c subcmd.Cmd, runctx *subcmd.RunContext) *subcmd.Error {
	return c.Run(runctx)
}

// Run is invoked by main(), each argument means 'volt {subcmd} {args}'.
func Run(args []string, cont RunnerFunc) *subcmd.Error {
	if os.Getenv("VOLT_DEBUG") != "" {
		logger.SetLevel(logger.DebugLevel)
	}

	if len(args) <= 1 {
		args = append(args, "help")
	}
	cmdname := args[1]
	args = args[2:]

	// Read config.toml
	cfg, err := config.Read()
	if err != nil {
		err = errors.New("could not read config.toml: " + err.Error())
		return &subcmd.Error{Code: 2, Msg: err.Error()}
	}

	// Expand subcommand alias
	cmdname, args, err = expandAlias(cmdname, args, cfg)
	if err != nil {
		return &subcmd.Error{Code: 1, Msg: err.Error()}
	}

	c, exists := subcmd.LookupSubcmd(cmdname)
	if !exists {
		return &subcmd.Error{Code: 3, Msg: "Unknown command '" + cmdname + "'"}
	}

	// Disallow executing the commands which may modify files in root priviledge
	if c.ProhibitRootExecution(args) {
		err := detectPriviledgedUser()
		if err != nil {
			return &subcmd.Error{Code: 4, Msg: err.Error()}
		}
	}

	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		err = errors.New("failed to read lock.json: " + err.Error())
		return &subcmd.Error{Code: 5, Msg: err.Error()}
	}

	return cont(c, &subcmd.RunContext{
		Args:     args,
		LockJSON: lockJSON,
		Config:   cfg,
	})
}

func expandAlias(cmdname string, args []string, cfg *config.Config) (string, []string, error) {
	if newArgs, exists := cfg.Alias[cmdname]; exists && len(newArgs) > 0 {
		cmdname = newArgs[0]
		args = append(newArgs[1:], args...)
	}
	return cmdname, args, nil
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
