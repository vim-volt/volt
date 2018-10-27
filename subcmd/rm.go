package subcmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/vim-volt/volt/fileutil"
	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/pathutil"
	"github.com/vim-volt/volt/plugconf"
	"github.com/vim-volt/volt/subcmd/builder"
	"github.com/vim-volt/volt/transaction"
)

func init() {
	cmdMap["rm"] = &rmCmd{}
}

type rmCmd struct {
	helped     bool
	rmRepos    bool
	rmPlugconf bool
}

func (cmd *rmCmd) ProhibitRootExecution(args []string) bool { return true }

func (cmd *rmCmd) FlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Print(`
Usage
  volt rm [-help] [-r] [-p] {repository} [{repository2} ...]

Quick example
  $ volt rm tyru/caw.vim    # Remove tyru/caw.vim plugin from lock.json
  $ volt rm -r tyru/caw.vim # Remove tyru/caw.vim plugin from lock.json, and remove repository directory
  $ volt rm -p tyru/caw.vim # Remove tyru/caw.vim plugin from lock.json, and remove plugconf
  $ volt rm -r -p tyru/caw.vim # Remove tyru/caw.vim plugin from lock.json, and remove repository directory, plugconf

Description
  Uninstall one or more {repository} from every profile.
  This results in removing vim plugins from ~/.vim/pack/volt/opt/ directory.
  If {repository} is depended by other repositories, this command exits with an error.

  If -r option was given, remove also repository directories of specified repositories.
  If -p option was given, remove also plugconf files of specified repositories.

  {repository} is treated as same format as "volt get" (see "volt get -help").` + "\n\n")
		//fmt.Println("Options")
		//fs.PrintDefaults()
		fmt.Println()
		cmd.helped = true
	}
	fs.BoolVar(&cmd.rmRepos, "r", false, "remove also repository directories")
	fs.BoolVar(&cmd.rmPlugconf, "p", false, "remove also plugconf files")
	return fs
}

func (cmd *rmCmd) Run(args []string) *Error {
	reposPathList, err := cmd.parseArgs(args)
	if err == ErrShowedHelp {
		return nil
	}
	if err != nil {
		return &Error{Code: 10, Msg: err.Error()}
	}

	err = cmd.doRemove(reposPathList)
	if err != nil {
		return &Error{Code: 11, Msg: "Failed to remove repository: " + err.Error()}
	}

	// Build opt dir
	err = builder.Build(false)
	if err != nil {
		return &Error{Code: 12, Msg: "Could not build " + pathutil.VimVoltDir() + ": " + err.Error()}
	}

	return nil
}

func (cmd *rmCmd) parseArgs(args []string) ([]pathutil.ReposPath, error) {
	fs := cmd.FlagSet()
	fs.Parse(args)
	if cmd.helped {
		return nil, ErrShowedHelp
	}

	if len(fs.Args()) == 0 {
		fs.Usage()
		return nil, errors.New("repository was not given")
	}

	var reposPathList []pathutil.ReposPath
	for _, arg := range fs.Args() {
		reposPath, err := pathutil.NormalizeRepos(arg)
		if err != nil {
			return nil, err
		}
		reposPathList = append(reposPathList, reposPath)
	}
	return reposPathList, nil
}

func (cmd *rmCmd) doRemove(reposPathList []pathutil.ReposPath) error {
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

	// Check if specified plugins are depended by some plugins
	for _, reposPath := range reposPathList {
		rdeps, err := plugconf.RdepsOf(reposPath, lockJSON.Repos)
		if err != nil {
			return err
		}
		if len(rdeps) > 0 {
			return errors.Errorf("cannot remove '%s' because it's depended by '%s'",
				reposPath, strings.Join(rdeps.Strings(), "', '"))
		}
	}

	removeCount := 0
	for _, reposPath := range reposPathList {
		// Remove repository directory
		if cmd.rmRepos {
			fullReposPath := reposPath.FullPath()
			if pathutil.Exists(fullReposPath) {
				if err = cmd.removeRepos(fullReposPath); err != nil {
					return err
				}
				removeCount++
			} else {
				logger.Debugf("No repository was installed for '%s' ... skip.", reposPath)
			}
		}

		// Remove plugconf file
		if cmd.rmPlugconf {
			plugconfPath := reposPath.Plugconf()
			if pathutil.Exists(plugconfPath) {
				if err = cmd.removePlugconf(plugconfPath); err != nil {
					return err
				}
				removeCount++
			} else {
				logger.Debugf("No plugconf was installed for '%s' ... skip.", reposPath)
			}
		}

		// Remove repository from lock.json
		err = lockJSON.Repos.RemoveAllReposPath(reposPath)
		err2 := lockJSON.Profiles.RemoveAllReposPath(reposPath)
		if err == nil || err2 == nil {
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
