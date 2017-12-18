package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/src-d/go-git.v4"

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
  volt add [{dir} ...]

Quick example
  $ mkdir -p ~/volt/repos/localhost/local/hello/plugin
  $ echo 'command! Hello echom "hello"' >~/volt/repos/localhost/local/hello/plugin/hello.vim
  $ volt add hello   # same as "volt add localhost/local/hello"
  $ vim -c Hello     # will output "hello"

Description
    Manage given local directories or git repositories as a vim plugin.
    This is useful in such cases:
    * Manage ~/.vim/* files as one repository
    * Manage unpublished plugins
    If {dir} does not contain any "/", it is treated as
    "localhost/local/{dir}" repository.
    If {dir} contains "/", it is treated as
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

	reposPathList, err := cmd.parseArgs(args)
	if err == ErrShowedHelp {
		return 0
	}
	if err != nil {
		logger.Error("Failed to parse args: " + err.Error())
		return 10
	}

	err = cmd.doAdd(reposPathList)
	if err != nil {
		logger.Error("Failed to add: " + err.Error())
		return 11
	}

	return 0
}

func (*addCmd) parseArgs(args []string) ([]string, error) {
	fs := cmdFlagSet["add"]
	fs.Parse(args)
	if addFlags.helped {
		return nil, ErrShowedHelp
	}
	fsArgs := fs.Args()
	if len(fsArgs) == 0 {
		fs.Usage()
		return nil, errors.New("invalid arguments")
	}
	reposPathList := make([]string, 0, len(fsArgs))
	for _, arg := range fsArgs {
		if strings.Contains(arg, "/") {
			reposPath, err := pathutil.NormalizeLocalRepos(arg)
			if err != nil {
				return nil, err
			}
			reposPathList = append(reposPathList, reposPath)
		} else {
			reposPathList = append(reposPathList, "localhost/local/"+arg)
		}
	}
	return reposPathList, nil
}

func (cmd *addCmd) doAdd(reposPathList []string) error {
	// Return an error if any errors are detected
	reposTypeList := make([]lockjson.ReposType, 0, len(reposPathList))
	for _, reposPath := range reposPathList {
		if reposType, err := cmd.checkRepos(reposPath); err != nil {
			return err
		} else {
			reposTypeList = append(reposTypeList, reposType)
		}
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

	// Find matching profile
	profile, err := lockJSON.Profiles.FindByName(lockJSON.CurrentProfileName)
	if err != nil {
		return err
	}

	// Add repos to repos list and current profile
	willUpdate := false
	for i, reposPath := range reposPathList {
		updated := false
		// Add repos to lockJSON
		if !lockJSON.Repos.Contains(reposPath) {
			lockJSON.Repos = append(lockJSON.Repos, lockjson.Repos{
				Type:    reposTypeList[i],
				TrxID:   lockJSON.TrxID,
				Path:    reposPath,
				Version: "",
			})
			updated = true
		}

		// Add repos to profiles[]/repos_path
		if !profile.ReposPath.Contains(reposPath) {
			// Add repos to 'profiles[]/repos_path'
			profile.ReposPath = append(profile.ReposPath, reposPath)
			updated = true
		}

		if updated {
			willUpdate = true
			logger.Infof("Added %s", reposPath)
		}
	}

	if willUpdate {
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
	}

	return nil
}

func (cmd *addCmd) checkRepos(reposPath string) (lockjson.ReposType, error) {
	fullpath := pathutil.FullReposPathOf(reposPath)
	// Return if any repositories do not exist
	if !pathutil.Exists(fullpath) {
		return "", errors.New("the repository does not exist: " + fullpath)
	}
	// Return if volt cannot detect repos type
	reposType, err := cmd.detectReposType(fullpath)
	if err != nil {
		return "", err
	}
	// Return if any irregular files exist
	err = filepath.Walk(fullpath, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.Mode()&BuildModeInvalidType != 0 {
			return errors.New(ErrBuildModeType + ": " + path)
		}
		return nil
	})
	return reposType, err
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
