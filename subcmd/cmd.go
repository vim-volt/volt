package subcmd

import (
	"flag"
	"os"
	"os/user"
	"runtime"

	"github.com/pkg/errors"

	"github.com/vim-volt/volt/config"
	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/logger"
)

var cmdMap = make(map[string]Cmd)

// Cmd represents volt's subcommand interface.
// All subcommands must implement this.
type Cmd interface {
	ProhibitRootExecution(args []string) bool
	Run(cmdctx *CmdContext) *Error
	FlagSet() *flag.FlagSet
}

// CmdContext is a data transfer object between Subcmd and Gateway layer.
type CmdContext struct {
	Args     []string
	LockJSON *lockjson.LockJSON
	Config   *config.Config
}

// RunnerFunc invokes c with args.
// On unit testing, a mock function was given.
type RunnerFunc func(c Cmd, cmdctx *CmdContext) *Error

// Error is a command error.
// It also has a exit code.
type Error struct {
	Code int
	Msg  string
}

func (e *Error) Error() string {
	return e.Msg
}

// DefaultRunner simply runs command with args
func DefaultRunner(c Cmd, cmdctx *CmdContext) *Error {
	return c.Run(cmdctx)
}

// Run is invoked by main(), each argument means 'volt {subcmd} {args}'.
func Run(args []string, cont RunnerFunc) *Error {
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
		return &Error{Code: 1, Msg: err.Error()}
	}

	c, exists := cmdMap[subCmd]
	if !exists {
		return &Error{Code: 3, Msg: "Unknown command '" + subCmd + "'"}
	}

	// Disallow executing the commands which may modify files in root priviledge
	if c.ProhibitRootExecution(args) {
		err := detectPriviledgedUser()
		if err != nil {
			return &Error{Code: 4, Msg: err.Error()}
		}
	}

	lockJSON, err := lockjson.Read()
	if err != nil {
		return &Error{Code: 20, Msg: errors.Wrap(err, "failed to read lock.json").Error()}
	}

	cfg, err := config.Read()
	if err != nil {
		return &Error{Code: 30, Msg: errors.Wrap(err, "failed to read config.toml").Error()}
	}

	return cont(c, &CmdContext{
		Args:     args,
		LockJSON: lockJSON,
		Config:   cfg,
	})
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
