package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/vim-volt/volt/fileutil"
	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/pathutil"
	"github.com/vim-volt/volt/transaction"
)

type addCmd struct{}

func Add(args []string) int {
	cmd := addCmd{}

	from, reposPath, err := cmd.parseArgs(args)
	if err != nil {
		logger.Error("Failed to parse args: " + err.Error())
		return 10
	}

	err = cmd.doAdd(from, reposPath)
	if err != nil {
		logger.Error("Failed to add: " + err.Error())
		return 11
	}

	return 0
}

func (*addCmd) parseArgs(args []string) (string, string, error) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Println(`
Usage
  volt add {from} {local repository}

Quick example
  $ mkdir -p hello/plugin
  $ echo 'command! Hello echom "hello"' >hello/plugin/hello.vim
  $ volt add hello
  $ vim -c Hello # will output "hello"

Description
    Add local {from} repository as {local repository} to lock.json .
    If {local repository} does not contain "/", it is treated as
    "localhost/local/{local repository}" repository.
    If {local repository} contains "/", it is treated as
    same format as "volt get" (see "volt get -help").

Options`)
		fs.PrintDefaults()
		fmt.Println()
	}
	fs.Parse(args)

	fsArgs := fs.Args()
	if len(fsArgs) == 2 {
		reposPath, err := pathutil.NormalizeLocalRepos(fsArgs[1])
		return fsArgs[0], reposPath, err
	} else {
		fs.Usage()
		return "", "", errors.New("invalid arguments")
	}
}

func (cmd *addCmd) doAdd(from, reposPath string) error {
	// Check from and destination (full path of repos path) path
	if !pathutil.Exists(from) {
		return errors.New("no such a directory: " + from)
	}

	dst := pathutil.FullReposPathOf(reposPath)
	if pathutil.Exists(dst) {
		return errors.New("the repository already exists: " + reposPath)
	}

	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		return errors.New("could not read lock.json: " + err.Error())
	}

	// Begin transaction
	err = transaction.Create()
	if err != nil {
		return errors.New("failed to begin transaction: " + err.Error())
	}
	defer transaction.Remove()
	lockJSON.TrxID++

	logger.Infof("Adding '%s' as '%s' ...", from, reposPath)

	// Copy directory from to dst
	err = fileutil.CopyDir(from, dst)
	if err != nil {
		return err
	}

	// Find matching profile
	profile, err := lockJSON.Profiles.FindByName(lockJSON.ActiveProfile)
	if err != nil {
		return err
	}

	// Add repos to lockJSON
	reposType, err := cmd.detectReposType(dst)
	lockJSON.Repos = append(lockJSON.Repos, lockjson.Repos{
		Type:  reposType,
		TrxID: lockJSON.TrxID,
		Path:  reposPath,
	})

	// Add repos to profiles[]/repos_path
	if !profile.ReposPath.Contains(reposPath) {
		// Add repos to 'profiles[]/repos_path'
		profile.ReposPath = append(profile.ReposPath, reposPath)
	}

	// Write to lock.json
	err = lockJSON.Write()
	if err != nil {
		return errors.New("could not write to lock.json: " + err.Error())
	}

	// Rebuild start dir
	err = (&rebuildCmd{}).doRebuild(false)
	if err != nil {
		return errors.New("could not rebuild " + pathutil.VimVoltDir() + ": " + err.Error())
	}

	return nil
}

func (*addCmd) detectReposType(fullpath string) (lockjson.ReposType, error) {
	if !pathutil.Exists(fullpath) {
		return "", errors.New("no such a directory: " + fullpath)
	}
	if pathutil.Exists(filepath.Join(fullpath, ".git")) {
		return lockjson.ReposGitType, nil
	}
	return lockjson.ReposStaticType, nil
}
