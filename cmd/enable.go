package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/vim-volt/go-volt/lockjson"
	"github.com/vim-volt/go-volt/pathutil"
	"github.com/vim-volt/go-volt/transaction"
)

type enableCmd struct{}

func Enable(args []string) int {
	cmd := enableCmd{}

	reposPath, err := cmd.parseArgs(args)
	if err != nil {
		fmt.Println("[ERROR] Failed to parse args: " + err.Error())
		return 10
	}

	err = cmd.setActive(reposPath, true)
	if err != nil {
		fmt.Println("[ERROR] Could not activate " + reposPath + ": " + err.Error())
		return 11
	}

	fmt.Println("[INFO] Activated " + reposPath)

	return 0
}

func (enableCmd) parseArgs(args []string) (string, error) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Println(`
Usage
  volt enable {repository}

Description
  Set active flag of {repository} to true
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

	return reposPath, nil
}

func (enableCmd) setActive(reposPath string, active bool) error {
	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		return errors.New("failed to read lock.json: " + err.Error())
	}

	// Find matching repos
	var repos *lockjson.Repos
	for i := range lockJSON.Repos {
		if lockJSON.Repos[i].Path == reposPath {
			repos = &lockJSON.Repos[i]
			break
		}
	}
	if repos == nil {
		return errors.New("no matching repos")
	}
	repos.Active = active

	// Begin transaction
	err = transaction.Create()
	if err != nil {
		return err
	}
	defer transaction.Remove()
	lockJSON.TrxID++

	// Write to lock.json
	err = lockjson.Write(lockJSON)
	if err != nil {
		return err
	}

	return nil
}
