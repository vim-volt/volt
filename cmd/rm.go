package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vim-volt/volt/fileutil"
	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/pathutil"
	"github.com/vim-volt/volt/plugconf"
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

	// Check if specified plugins are depended by some plugins
	for _, reposPath := range reposPathList {
		rdeps, err := plugconf.RdepsOf(reposPath, lockJSON.Repos)
		if err != nil {
			return err
		}
		if len(rdeps) > 0 {
			return fmt.Errorf("cannot remove '%s' because it's depended by '%s'",
				reposPath, strings.Join(rdeps, "', '"))
		}
	}

	// Remove each repository
	removeCount := 0
	for _, reposPath := range reposPathList {
		fullReposPath := pathutil.FullReposPathOf(reposPath)
		// Remove repository directory
		if pathutil.Exists(fullReposPath) {
			if err = cmd.removeRepos(fullReposPath); err != nil {
				return err
			}
			removeCount++
		} else {
			logger.Debugf("No repository was installed for '%s' ... skip.", reposPath)
		}
		// Remove plugconf file
		if plugconfPath := pathutil.PlugconfOf(reposPath); flags.plugconf && pathutil.Exists(plugconfPath) {
			if err = cmd.removePlugconf(plugconfPath); err != nil {
				return err
			}
			removeCount++
		} else {
			logger.Debugf("No plugconf was installed for '%s' ... skip.", reposPath)
		}
		// Update lockJSON
		if lockJSON.Repos.RemoveAllByPath(reposPath) == nil ||
			lockJSON.Profiles.RemoveAllReposPath(reposPath) == nil {
			removeCount++
		}
	}
	if removeCount == 0 {
		return errors.New("no plugins are removed")
	}

	// Write to lock.json
	if err = lockJSON.Write(); err != nil {
		return err
	}
	return nil
}

// Remove repository directory
func (cmd *rmCmd) removeRepos(fullReposPath string) error {
	logger.Info("Removing " + fullReposPath + " ...")
	if err := os.RemoveAll(fullReposPath); err != nil {
		return err
	}
	fileutil.RemoveDirs(filepath.Dir(fullReposPath))
	return nil
}

// Remove plugconf file
func (*rmCmd) removePlugconf(plugconfPath string) error {
	logger.Info("Removing plugconf files ...")
	if err := os.Remove(plugconfPath); err != nil {
		return err
	}
	// Remove parent directories of plugconf
	fileutil.RemoveDirs(filepath.Dir(plugconfPath))
	return nil
}
