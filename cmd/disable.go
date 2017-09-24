package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/vim-volt/go-volt/pathutil"
)

type disableCmd struct{}

func Disable(args []string) int {
	cmd := disableCmd{}

	reposPath, err := cmd.parseArgs(args)
	if err != nil {
		fmt.Println("[ERROR] Failed to parse args: " + err.Error())
		return 10
	}

	err = enableCmd{}.setActive(reposPath, false)
	if err != nil {
		fmt.Println("[ERROR] Could not deactivate " + reposPath + ": " + err.Error())
		return 11
	}

	return 0
}

func (disableCmd) parseArgs(args []string) (string, error) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Println(`
Usage
  volt disable {repository}

Description
  Set active flag of {repository} to false
  to be determined if vim-volt loads it or not.

Options`)
		fs.PrintDefaults()
		fmt.Println()
	}
	fs.Parse(args)

	if len(fs.Args()) == 0 {
		fs.Usage()
		return "", errors.New("repository was not given")
	}

	// Normalize repos path
	reposPath, err := pathutil.NormalizeRepository(fs.Args()[0])
	if err != nil {
		return "", err
	}

	fmt.Println("[INFO] Deactivated " + reposPath)

	return reposPath, nil
}
