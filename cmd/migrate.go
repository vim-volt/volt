package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/vim-volt/volt/cmd/migrate"
	"github.com/vim-volt/volt/logger"
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
		fmt.Print(`
Usage
  volt migrate [-help] {migration operation}

Description
    Perform migration of $VOLTPATH/lock.json, which means volt converts old version lock.json structure into the latest version. This is always done automatically when reading lock.json content. For example, 'volt get <repos>' will install plugin, and migrate lock.json structure, and write it to lock.json after all. so the migrated content is written to lock.json automatically.
    But, for example, 'volt list' does not write to lock.json but does read, so every time when running 'volt list' shows warning about lock.json is old.
    To suppress this, running this command simply reads and writes migrated structure to lock.json.` + "\n\n")
		//fmt.Println("Options")
		//fs.PrintDefaults()
		fmt.Println()
		cmd.helped = true
	}
	return fs
}

func (cmd *migrateCmd) Run(args []string) int {
	op, err := cmd.parseArgs(args)
	if err == ErrShowedHelp {
		return 0
	}
	if err != nil {
		logger.Error("Failed to parse args: " + err.Error())
		cmd.showAvailableOps()
		return 10
	}

	if err := op.Migrate(); err != nil {
		logger.Error("Failed to migrate: " + err.Error())
		return 11
	}

	logger.Infof("'%s' was successfully migrated!", op.Name())
	return 0
}

func (cmd *migrateCmd) parseArgs(args []string) (migrate.Migrater, error) {
	fs := cmd.FlagSet()
	fs.Parse(args)
	if cmd.helped {
		return nil, ErrShowedHelp
	}
	if len(args) == 0 {
		return nil, errors.New("please specify migration operation")
	}
	return migrate.GetMigrater(args[0])
}

func (cmd *migrateCmd) showAvailableOps() {
	logger.Info("Available migrate operations are:")
	for _, m := range migrate.ListMigraters() {
		logger.Infof("  %s - %s", m.Name(), m.Description())
	}
}
