package gateway

import (
	"flag"
	"fmt"
	"os"

	"github.com/vim-volt/volt/usecase"
)

func init() {
	cmdMap["version"] = &versionCmd{Version: usecase.VersionString()}
}

type versionCmd struct {
	helped bool

	Version string
}

func (cmd *versionCmd) ProhibitRootExecution(args []string) bool { return false }

func (cmd *versionCmd) FlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Print(`
Usage
  volt version [-help]

Description
  Show current version of volt.` + "\n\n")
		//fmt.Println("Options")
		//fs.PrintDefaults()
		fmt.Println()
		cmd.helped = true
	}
	return fs
}

func (cmd *versionCmd) Run(cmdctx *CmdContext) *Error {
	fs := cmd.FlagSet()
	fs.Parse(cmdctx.Args)
	if cmd.helped {
		return nil
	}

	fmt.Printf("volt version: %s\n", cmd.Version)
	return nil
}
