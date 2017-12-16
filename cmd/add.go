package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/src-d/go-git.v4"

	"github.com/vim-volt/volt/fileutil"
	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/pathutil"
	"github.com/vim-volt/volt/transaction"
)

type addFlagsType struct {
	helped bool
}

var addFlags addFlagsType

func init() {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Print(`
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
    same format as "volt get" (see "volt get -help").` + "\n\n")
		//fmt.Println("Options")
		//fs.PrintDefaults()
		fmt.Println()
		addFlags.helped = true
	}

	cmdFlagSet["add"] = fs
}

type addCmd struct{}

func Add(args []string) int {
	cmd := addCmd{}

	from, reposPath, err := cmd.parseArgs(args)
	if err == ErrShowedHelp {
		return 0
	}
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
	fs := cmdFlagSet["add"]
	fs.Parse(args)
	if addFlags.helped {
		return "", "", ErrShowedHelp
	}

	fsArgs := fs.Args()
	if len(fsArgs) == 1 {
		reposPath := "localhost/local/" + filepath.Base(fsArgs[0])
		return fsArgs[0], reposPath, nil
	} else if len(fsArgs) == 2 {
		reposPath, err := pathutil.NormalizeLocalRepos(fsArgs[1])
		return fsArgs[0], reposPath, err
	} else {
		fs.Usage()
		return "", "", errors.New("invalid arguments")
	}
}

func (cmd *addCmd) doAdd(from, reposPath string) error {
	fromInfo, err := os.Stat(from)
	if err != nil {
		return err
	} else if !fromInfo.IsDir() {
		return fmt.Errorf("source is not a directory: " + from)
	}

	dst := pathutil.FullReposPathOf(reposPath)
	if pathutil.Exists(dst) {
		return errors.New("the repository already exists: " + reposPath)
	}

	reposType, err := cmd.detectReposType(dst)
	if err != nil {
		return err
	}

	// Check if regular files exist
	err = filepath.Walk(from, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.Mode()&BuildModeInvalidType != 0 {
			return errors.New(ErrBuildModeType + ": " + path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		return errors.New("could not read lock.json: " + err.Error())
	}

	// Begin transaction
	err = transaction.Create()
	if err != nil {
		return err
	}
	defer transaction.Remove()
	lockJSON.TrxID++

	logger.Infof("Adding '%s' as '%s' ...", from, reposPath)

	// Copy directory from to dst
	buf := make([]byte, 32*1024)
	err = fileutil.CopyDir(from, dst, buf, fromInfo.Mode(), BuildModeInvalidType)
	if err != nil {
		if e, ok := err.(*fileutil.InvalidTypeError); ok {
			return errors.New(ErrBuildModeType + ": " + e.Filename)
		}
		return err
	}

	// Find matching profile
	profile, err := lockJSON.Profiles.FindByName(lockJSON.CurrentProfileName)
	if err != nil {
		return err
	}

	// Add repos to lockJSON
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

	// Build ~/.vim/pack/volt dir
	err = (&buildCmd{}).doBuild(false)
	if err != nil {
		return errors.New("could not build " + pathutil.VimVoltDir() + ": " + err.Error())
	}

	return nil
}

func (*addCmd) detectReposType(fullpath string) (lockjson.ReposType, error) {
	if pathutil.Exists(filepath.Join(fullpath, ".git")) {
		if _, err := git.PlainOpen(fullpath); err != nil {
			return "", err
		}
		return lockjson.ReposGitType, nil
	}
	return lockjson.ReposStaticType, nil
}
