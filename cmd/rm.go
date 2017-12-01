package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/pathutil"
	"github.com/vim-volt/volt/transaction"
)

type rmFlagsType struct {
	helped   bool
	plugconf bool
}

var rmFlags rmFlagsType

func init() {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Print(`
Usage
  volt rm [-help] [-p] {repository} [{repository2} ...]

Quick example
  $ volt rm tyru/caw.vim    # Uninstall tyru/caw.vim plugin
  $ volt rm -p tyru/caw.vim # Uninstall tyru/caw.vim plugin and plugconf file

Description
  Uninstall vim plugin of {repository} on every profile.
  If -p option was given, remove also plugconf files of specified plugins.

  {repository} is treated as same format as "volt get" (see "volt get -help").` + "\n\n")
		//fmt.Println("Options")
		//fs.PrintDefaults()
		fmt.Println()
		rmFlags.helped = true
	}
	fs.BoolVar(&rmFlags.plugconf, "p", false, "remove also plugconf file")

	cmdFlagSet["rm"] = fs
}

type rmCmd struct{}

func Rm(args []string) int {
	cmd := rmCmd{}

	reposPathList, flags, err := cmd.parseArgs(args)
	if err == ErrShowedHelp {
		return 0
	}
	if err != nil {
		logger.Error(err.Error())
		return 10
	}

	err = cmd.doRemove(reposPathList, flags)
	if err != nil {
		logger.Error("Failed to remove repository: " + err.Error())
		return 11
	}

	// Rebuild opt dir
	err = (&rebuildCmd{}).doRebuild(false)
	if err != nil {
		logger.Error("could not rebuild " + pathutil.VimVoltDir() + ": " + err.Error())
		return 12
	}

	return 0
}

func (*rmCmd) parseArgs(args []string) ([]string, *rmFlagsType, error) {
	fs := cmdFlagSet["rm"]
	fs.Parse(args)
	if rmFlags.helped {
		return nil, nil, ErrShowedHelp
	}

	if len(fs.Args()) == 0 {
		fs.Usage()
		return nil, nil, errors.New("repository was not given")
	}

	var reposPathList []string
	for _, arg := range fs.Args() {
		reposPath, err := pathutil.NormalizeRepos(arg)
		if err != nil {
			return nil, nil, err
		}
		reposPathList = append(reposPathList, reposPath)
	}
	return reposPathList, &rmFlags, nil
}

func (cmd *rmCmd) doRemove(reposPathList []string, flags *rmFlagsType) error {
	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		return err
	}

	// Begin transaction
	err = transaction.Create()
	if err != nil {
		return err
	}
	defer transaction.Remove()
	lockJSON.TrxID++

	// Remove each repository
	for _, reposPath := range reposPathList {
		// Remove repository directory
		err = cmd.removeRepos(reposPath)
		if err != nil {
			return err
		}
		if flags.plugconf {
			// Remove plugconf file
			err = cmd.removePlugconf(reposPath)
			if err != nil {
				return err
			}
		}
		// Update lockJSON
		err = lockJSON.Repos.RemoveAllByPath(reposPath)
		if err != nil {
			return err
		}
		lockJSON.Profiles.RemoveAllReposPath(reposPath)
	}

	// Write to lock.json
	return lockJSON.Write()
}

// Remove repository directory
func (cmd *rmCmd) removeRepos(reposPath string) error {
	fullpath := pathutil.FullReposPathOf(reposPath)
	logger.Info("Removing " + fullpath + " ...")
	if pathutil.Exists(fullpath) {
		err := os.RemoveAll(fullpath)
		if err != nil {
			return err
		}
		cmd.removeDirs(filepath.Dir(fullpath))
	} else {
		return errors.New("no repository was installed: " + fullpath)
	}

	return nil
}

// Remove plugconf file
func (cmd *rmCmd) removePlugconf(reposPath string) error {
	logger.Info("Removing plugconf files ...")
	plugconf := pathutil.PlugconfOf(reposPath)
	if pathutil.Exists(plugconf) {
		err := os.Remove(plugconf)
		if err != nil {
			return err
		}
	}
	// Remove parent directories of plugconf
	cmd.removeDirs(filepath.Dir(plugconf))
	return nil
}

// Always returns non-nil error which is the last error of os.Remove(dir)
func (cmd *rmCmd) removeDirs(dir string) error {
	// Remove trailing slashes
	dir = strings.TrimRight(dir, "/")

	if err := os.Remove(dir); err != nil {
		return err
	} else {
		return cmd.removeDirs(filepath.Dir(dir))
	}
}
