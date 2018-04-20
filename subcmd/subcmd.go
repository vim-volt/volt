package subcmd

import (
	"flag"

	"github.com/vim-volt/volt/config"
	"github.com/vim-volt/volt/lockjson"
)

var cmdMap = make(map[string]Cmd)

// LookupSubcmd looks up subcommand name
func LookupSubcmd(name string) (cmd Cmd, exists bool) {
	cmd, exists = cmdMap[name]
	return
}

// Cmd represents volt's subcommand interface.
// All subcommands must implement this.
type Cmd interface {
	ProhibitRootExecution(args []string) bool
	Run(runctx *RunContext) *Error
	FlagSet() *flag.FlagSet
}

// RunContext is a struct to have data which are passed between gateway package
// and subcmd package
type RunContext struct {
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
