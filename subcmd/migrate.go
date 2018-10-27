package subcmd

import (
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"os"

	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/subcmd/migrate"
)

func init() {
	cmdMap["migrate"] = &migrateCmd{}
}

type migrateCmd struct {
	helped bool
}

func (cmd *migrateCmd) ProhibitRootExecution(args []string) bool { return true }

func (cmd *migrateCmd) FlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		args := fs.Args()
		if len(args) > 0 {
			m, err := migrate.GetMigrater(args[0])
			if err != nil {
				return
			}
			fmt.Println(m.Description(false))
			fmt.Println()
			cmd.helped = true
			return
		}

		fmt.Println(`Usage
  volt migrate [-help] {migration operation}

Description
  Perform miscellaneous migration operations.
  See detailed help for 'volt migrate -help {migration operation}'.

Available operations`)
		cmd.showAvailableOps(func(line string) {
			fmt.Println(line)
		})
		//fmt.Println("Options")
		//fs.PrintDefaults()
		fmt.Println()
		cmd.helped = true
	}
	return fs
}

func (cmd *migrateCmd) Run(args []string) *Error {
	op, err := cmd.parseArgs(args)
	if err == ErrShowedHelp {
		return nil
	}
	if err != nil {
		return &Error{Code: 10, Msg: "Failed to parse args: " + err.Error()}
	}

	if err := op.Migrate(); err != nil {
		return &Error{Code: 11, Msg: "Failed to migrate: " + err.Error()}
	}

	logger.Infof("'%s' was successfully migrated!", op.Name())
	return nil
}

func (cmd *migrateCmd) parseArgs(args []string) (migrate.Migrater, error) {
	fs := cmd.FlagSet()
	fs.Parse(args)
	if cmd.helped {
		return nil, ErrShowedHelp
	}
	args = fs.Args()
	if len(args) == 0 {
		return nil, errors.New("please specify migration operation")
	}
	return migrate.GetMigrater(args[0])
}

func (cmd *migrateCmd) showAvailableOps(write func(string)) {
	for _, m := range migrate.ListMigraters() {
		write(fmt.Sprintf("  %s", m.Name()))
		write(fmt.Sprintf("    %s", m.Description(true)))
	}
}
