package gateway

import (
	"flag"

	"github.com/vim-volt/volt/config"
	"github.com/vim-volt/volt/lockjson"
)

var cmdMap = make(map[string]Cmd)

// LookUpCmd looks up subcommand by name.
func LookUpCmd(cmd string) Cmd {
	return cmdMap[cmd]
}

// Cmd represents volt's subcommand interface.
// All subcommands must implement this.
type Cmd interface {
	ProhibitRootExecution(args []string) bool
	Run(cmdctx *CmdContext) *Error
	FlagSet() *flag.FlagSet
}

// CmdContext is a data transfer object between Subcmd and Gateway layer.
type CmdContext struct {
	Cmd      string
	Args     []string
	LockJSON *lockjson.LockJSON
	Config   *config.Config
}

// Error is a command error.
// It also has a exit code.
type Error struct {
	Code int
	Msg  string
}

func (e *Error) Error() string {
	return e.Msg
}
