package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"text/template"

	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/logger"
)

type inspectFlagsType struct {
	helped bool
	format string
}

var inspectFlags inspectFlagsType

func init() {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Print(`
Usage
  volt inspect -f {format}

Quick example
  This shows all installed repositories.

  $ volt inspect -f '{{ range .Repos }}{{ .Name }}{{ end }}'

  This shows repositories enabled by current profile.

  $ volt inspect -f '{{ range .CurrentProfile.ReposPath }}{{ . }}{{ end }}'

Description
  Shows internal information (plugins, profiles, ...) by given format.

    TODO

    ` + "\n\n")
		//fmt.Println("Options")
		//fs.PrintDefaults()
		fmt.Println()
		inspectFlags.helped = true
	}
	fs.StringVar(&inspectFlags.format, "format", "", "text/template format string")
	fs.StringVar(&inspectFlags.format, "f", "", "text/template format string")

	cmdFlagSet["inspect"] = fs
}

type inspectCmd struct{}

func Inspect(args []string) int {
	cmd := inspectCmd{}

	flags, err := cmd.parseArgs(args)
	if err == ErrShowedHelp {
		return 0
	}
	if err != nil {
		logger.Error("Failed to parse args: " + err.Error())
		return 10
	}

	err = cmd.doInspect(flags.format)
	if err != nil {
		logger.Error("Failed to inspect: " + err.Error())
		return 11
	}

	return 0
}

func (*inspectCmd) parseArgs(args []string) (*inspectFlagsType, error) {
	fs := cmdFlagSet["inspect"]
	fs.Parse(args)
	if inspectFlags.helped {
		return nil, ErrShowedHelp
	}

	if inspectFlags.format == "" {
		return nil, errors.New("-format option is required")
	}
	return &inspectFlags, nil
}

func (cmd *inspectCmd) doInspect(format string) error {
	// Parse template string
	t, err := template.New("volt").Parse(format)
	if err != nil {
		return err
	}

	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		return errors.New("could not read lock.json: " + err.Error())
	}

	// Output templated information
	err = t.Execute(os.Stdout, lockJSON)
	if err != nil {
		return err
	}

	return nil
}
