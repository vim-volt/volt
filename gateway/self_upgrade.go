package gateway

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/vim-volt/volt/usecase"
)

func init() {
	cmdMap["self-upgrade"] = &selfUpgradeCmd{
		SelfUpgrade:     usecase.SelfUpgrade,
		RemoveOldBinary: usecase.RemoveOldBinary,
	}
}

type selfUpgradeCmd struct {
	helped    bool
	checkOnly bool

	SelfUpgrade     func(latestURL string, checkOnly bool) error
	RemoveOldBinary func(ppid int) error
}

func (cmd *selfUpgradeCmd) ProhibitRootExecution(args []string) bool { return true }

func (cmd *selfUpgradeCmd) FlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Print(`
Usage
  volt self-upgrade [-help] [-check]

Description
    Upgrade to the latest volt command, or if -check was given, it only checks the newer version is available.` + "\n\n")
		//fmt.Println("Options")
		//fs.PrintDefaults()
		fmt.Println()
		cmd.helped = true
	}
	fs.BoolVar(&cmd.checkOnly, "check", false, "only checks the newer version is available")
	return fs
}

func (cmd *selfUpgradeCmd) Run(cmdctx *CmdContext) *Error {
	err := cmd.parseArgs(cmdctx.Args)
	if err == ErrShowedHelp {
		return nil
	}
	if err != nil {
		return &Error{Code: 10, Msg: "Failed to parse args: " + err.Error()}
	}

	if ppidStr := os.Getenv("VOLT_SELF_UPGRADE_PPID"); ppidStr != "" {
		ppid, err := strconv.Atoi(ppidStr)
		if err != nil {
			return &Error{Code: 20, Msg: "Failed to parse VOLT_SELF_UPGRADE_PPID: " + err.Error()}
		}
		if err = cmd.RemoveOldBinary(ppid); err != nil {
			return &Error{Code: 11, Msg: "Failed to clean up old binary: " + err.Error()}
		}
	} else {
		latestURL := "https://api.github.com/repos/vim-volt/volt/releases/latest"
		if err = cmd.SelfUpgrade(latestURL, cmd.checkOnly); err != nil {
			return &Error{Code: 12, Msg: "Failed to self-upgrade: " + err.Error()}
		}
	}

	return nil
}

func (cmd *selfUpgradeCmd) parseArgs(args []string) error {
	fs := cmd.FlagSet()
	fs.Parse(args)
	if cmd.helped {
		return ErrShowedHelp
	}
	return nil
}
