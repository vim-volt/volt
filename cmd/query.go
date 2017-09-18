package cmd

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/vim-volt/go-volt/lockjson"
	"github.com/vim-volt/go-volt/pathutil"
)

type queryCmd struct{}

type queryFlags struct {
	json      bool
	installed bool
}

func Query(args []string) int {
	cmd := queryCmd{}

	reposPath, flags, err := cmd.parseArgs(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 10
	}

	err = cmd.queryRepos(reposPath, flags)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to clone repository: "+err.Error())
		return 11
	}

	return 0
}

func (queryCmd) parseArgs(args []string) (string, *queryFlags, error) {
	var flags queryFlags
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `
Usage
  volt query [-help] [-j] [-i] [{repository}]

Description
  Output queried vim plugin info

Options`)
		fs.PrintDefaults()
		fmt.Fprintln(os.Stderr)
	}
	fs.BoolVar(&flags.json, "j", false, "output as JSON")
	fs.BoolVar(&flags.installed, "i", false, "show installed info")
	fs.Parse(args)

	if !flags.installed && len(fs.Args()) == 0 {
		fs.Usage()
		return "", nil, errors.New("repository was not given")
	}

	var reposPath string
	if len(fs.Args()) > 0 {
		var err error
		reposPath, err = pathutil.NormalizeRepository(fs.Args()[0])
		if err != nil {
			return "", nil, err
		}
	}
	return reposPath, &flags, nil
}

func (queryCmd) queryRepos(reposPath string, flags *queryFlags) error {
	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		return err
	}

	if !flags.json { // TODO
		return errors.New("specify -j option (showing non-JSON output is not supported)")
	}
	if !flags.installed { // TODO
		return errors.New("specify -i option (showing non-installed plugin info is not supported)")
	}

	if flags.json && flags.installed {
		bytes, err := json.Marshal(lockJSON.Repos)
		if err != nil {
			return err
		}
		fmt.Print(string(bytes))
	}

	return nil
}
