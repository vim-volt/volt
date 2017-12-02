package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/pathutil"
	"github.com/vim-volt/volt/transaction"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/protocol/packp/sideband"
)

type getFlagsType struct {
	helped   bool
	lockJSON bool
	upgrade  bool
	verbose  bool
}

var getFlags getFlagsType

func init() {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Println(`
Usage
  volt get [-help] [-l] [-u] [-v] [{repository} ...]

Quick example
  $ volt get tyru/caw.vim    # will install tyru/caw.vim plugin
  $ volt get -u tyru/caw.vim # will upgrade tyru/caw.vim plugin
  $ volt get -l -u           # will upgrade all installed plugins
  $ volt get -v tyru/caw.vim # will output verbose git-clone(1) output

Description
  Install vim plugin from {repository}, or upgrade vim plugin of {repository} list on current active profile. And fetch skeleton plugconf from:
    https://github.com/vim-volt/plugconf-templates
  and install it to:
    $VOLTPATH/plugconf/{repository}.vim

  {repository}'s format is one of the followings:

  1. {user}/{name}
       This is same as "github.com/{user}/{name}"
  2. {site}/{user}/{name}
  3. https://{site}/{user}/{name}
  4. http://{site}/{user}/{name}

  {repository} list is determined as followings:

  * If -l option is specified, all installed vim plugins (regardless current profile) are used
  * If {repository} arguments are specified, the specified vim plugins are used

  If -u is specified, upgrade given git repositories (static repositories are ignored).
  If -l option is specified, all installed vim plugins are used for targets to install or upgrade.
  If -l and -u options were specified together, upgrade all installed vim plugins (static repositories are ignored).

  If -v option was specified, show git-clone(1) output too.

Options`)
		fs.PrintDefaults()
		fmt.Println()
		getFlags.helped = true
	}
	fs.BoolVar(&getFlags.lockJSON, "l", false, "from lock.json")
	fs.BoolVar(&getFlags.upgrade, "u", false, "upgrade installed vim plugin")
	fs.BoolVar(&getFlags.verbose, "v", false, "show git-clone output")

	cmdFlagSet["get"] = fs
}

type getCmd struct{}

func Get(args []string) int {
	cmd := getCmd{}

	// Parse args
	args, flags, err := cmd.parseArgs(args)
	if err == ErrShowedHelp {
		return 0
	}
	if err != nil {
		logger.Error("Failed to parse args: " + err.Error())
		return 10
	}

	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		logger.Error("Could not read lock.json: " + err.Error())
		return 11
	}

	reposPathList, err := cmd.getReposPathList(flags, args, lockJSON)
	if err != nil {
		logger.Error("Could not get repos list: " + err.Error())
		return 12
	}

	err = cmd.doGet(reposPathList, flags, lockJSON)
	if err != nil {
		logger.Error(err.Error())
		return 13
	}

	return 0
}

func (*getCmd) parseArgs(args []string) ([]string, *getFlagsType, error) {
	fs := cmdFlagSet["get"]
	fs.Parse(args)
	if getFlags.helped {
		return nil, nil, ErrShowedHelp
	}

	if !getFlags.lockJSON && len(fs.Args()) == 0 {
		fs.Usage()
		return nil, nil, errors.New("repository was not given")
	}

	return fs.Args(), &getFlags, nil
}

func (*getCmd) getReposPathList(flags *getFlagsType, args []string, lockJSON *lockjson.LockJSON) ([]string, error) {
	reposPathList := make([]string, 0, 32)
	if flags.lockJSON {
		for _, repos := range lockJSON.Repos {
			reposPathList = append(reposPathList, repos.Path)
		}
	} else {
		for _, arg := range args {
			reposPath, err := pathutil.NormalizeRepos(arg)
			if err != nil {
				return nil, err
			}
			reposPathList = append(reposPathList, reposPath)
		}
	}
	return reposPathList, nil
}

func (cmd *getCmd) doGet(reposPathList []string, flags *getFlagsType, lockJSON *lockjson.LockJSON) error {
	// Find matching profile
	profile, err := lockJSON.Profiles.FindByName(lockJSON.ActiveProfile)
	if err != nil {
		// this must not be occurred because lockjson.Read()
		// validates if the matching profile exists
		return err
	}

	// Begin transaction
	err = transaction.Create()
	if err != nil {
		return err
	}
	defer transaction.Remove()
	lockJSON.TrxID++

	// Invoke installing / upgrading tasks
	done := make(chan getParallelResult)
	getCount := 0
	for _, reposPath := range reposPathList {
		repos, err := lockJSON.Repos.FindByPath(reposPath)
		if err != nil {
			repos = nil
		}
		if repos == nil || repos.Type == lockjson.ReposGitType {
			go cmd.getParallel(reposPath, repos, flags, done)
			getCount++
		}
	}

	// Wait results
	var statusList []string
	var updatedLockJSON bool
	for i := 0; i < getCount; i++ {
		r := <-done
		statusList = append(statusList, r.status)
		// Update repos[]/trx_id, repos[]/version
		if strings.HasPrefix(r.status, statusPrefixInstalled) ||
			strings.HasPrefix(r.status, statusPrefixUpgraded) {
			cmd.updateReposVersion(lockJSON, r.reposPath, r.hash, profile)
			updatedLockJSON = true
		}
	}

	if updatedLockJSON {
		// Write to lock.json
		err = lockJSON.Write()
		if err != nil {
			return errors.New("could not write to lock.json: " + err.Error())
		}

		// Rebuild ~/.vim/pack/volt dir
		err = (&rebuildCmd{}).doRebuild(false)
		if err != nil {
			return errors.New("could not rebuild " + pathutil.VimVoltDir() + ": " + err.Error())
		}
	}

	// Show results
	if len(statusList) > 0 {
		fmt.Print("\nDone!\n\n")
		for i := range statusList {
			fmt.Println(statusList[i])
		}
	}
	return nil
}

type getParallelResult struct {
	reposPath string
	status    string
	hash      string
}

const (
	statusPrefixFailed    = "!"
	statusPrefixNoChange  = "#"
	statusPrefixInstalled = "+"
	statusPrefixUpgraded  = "*"
)

// This function is executed in goroutine of each plugin
func (cmd *getCmd) getParallel(reposPath string, repos *lockjson.Repos, flags *getFlagsType, done chan getParallelResult) {
	// Normally, when upgraded is true, repos is also non-nil.
	var fromHash string
	if flags.upgrade && pathutil.Exists(pathutil.FullReposPathOf(reposPath)) {
		// Get HEAD hash string
		var err error
		fromHash, err = getReposHEAD(reposPath)
		if err != nil {
			logger.Error("Failed to get HEAD commit hash: " + err.Error())
			done <- getParallelResult{
				reposPath: reposPath,
				status:    fmt.Sprintf("%s %s : install failed", statusPrefixFailed, reposPath),
			}
			return
		}
	}

	var status string
	upgraded := false

	if flags.upgrade && pathutil.Exists(pathutil.FullReposPathOf(reposPath)) {
		// Upgrade plugin
		err := cmd.upgradePlugin(reposPath, flags)
		if err != git.NoErrAlreadyUpToDate && err != nil {
			logger.Warn("Failed to upgrade plugin: " + err.Error())

			done <- getParallelResult{
				reposPath: reposPath,
				status:    fmt.Sprintf("%s %s : upgrade failed : %s", statusPrefixFailed, reposPath, err.Error()),
			}
			return
		}
		if err == git.NoErrAlreadyUpToDate {
			status = fmt.Sprintf("%s %s : no change", statusPrefixNoChange, reposPath)
		} else {
			upgraded = true
		}
	} else {
		// Install plugin
		err := cmd.installPlugin(reposPath, flags)
		if err != nil {
			logger.Warn("Failed to install plugin: " + err.Error())
			done <- getParallelResult{
				reposPath: reposPath,
				status:    fmt.Sprintf("%s %s : install failed", statusPrefixFailed, reposPath),
			}
			return
		}
		status = fmt.Sprintf("%s %s : installed", statusPrefixInstalled, reposPath)

		// Install plugconf
		logger.Info("Installing plugconf " + reposPath + " ...")
		err = cmd.installPlugconf(reposPath)
		if err != nil {
			logger.Info("Installing plugconf " + reposPath + " ... not found")
		} else {
			logger.Info("Installing plugconf " + reposPath + " ... found")
		}
	}

	// Get HEAD hash string
	toHash, err := getReposHEAD(reposPath)
	if err != nil {
		logger.Error("Failed to get HEAD commit hash: " + err.Error())
		done <- getParallelResult{
			reposPath: reposPath,
			status:    fmt.Sprintf("%s %s : install failed", statusPrefixFailed, reposPath),
		}
		return
	}

	// Show old and new revisions: "upgraded ({from}..{to})".
	if upgraded && repos != nil {
		status = fmt.Sprintf("%s %s : upgraded (%s..%s)", statusPrefixUpgraded, reposPath, fromHash, toHash)
	}

	done <- getParallelResult{
		reposPath: reposPath,
		status:    status,
		hash:      toHash,
	}
}

func (cmd *getCmd) upgradePlugin(reposPath string, flags *getFlagsType) error {
	fullpath := pathutil.FullReposPathOf(reposPath)

	logger.Info("Upgrading " + reposPath + " ...")

	var progress sideband.Progress = nil
	if flags.verbose {
		progress = os.Stdout
	}

	repos, err := git.PlainOpen(fullpath)
	if err != nil {
		return err
	}

	cfg, err := repos.Config()
	if err != nil {
		return err
	}

	if cfg.Core.IsBare {
		return repos.Fetch(&git.FetchOptions{
			RemoteName: "origin",
			Progress:   progress,
		})
	} else {
		wt, err := repos.Worktree()
		if err != nil {
			return err
		}
		return wt.Pull(&git.PullOptions{
			RemoteName: "origin",
			Progress:   progress,
		})
	}
}

func (cmd *getCmd) installPlugin(reposPath string, flags *getFlagsType) error {
	fullpath := pathutil.FullReposPathOf(reposPath)
	if pathutil.Exists(fullpath) {
		return errors.New("repository exists")
	}

	logger.Info("Installing " + reposPath + " ...")

	var progress sideband.Progress = nil
	if flags.verbose {
		progress = os.Stdout
	}

	// Create parent directories
	err := os.MkdirAll(filepath.Dir(fullpath), 0755)
	if err != nil {
		return err
	}

	// Clone repository to $VOLTPATH/repos/{site}/{user}/{name}
	isBare := false
	_, err = git.PlainClone(fullpath, isBare, &git.CloneOptions{
		URL:      pathutil.CloneURLOf(reposPath),
		Progress: progress,
	})
	return err
}

func (cmd *getCmd) installPlugconf(reposPath string) error {
	// If non-nil error returned from FetchPlugconf(),
	// create skeleton plugconf file
	tmpl, _ := FetchPlugconf(reposPath)
	filename := pathutil.PlugconfOf(reposPath)
	content, err := GenPlugconfByTemplate(tmpl, filename)
	if err != nil {
		return err
	}
	os.MkdirAll(filepath.Dir(filename), 0755)
	err = ioutil.WriteFile(filename, content, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (*getCmd) updateReposVersion(lockJSON *lockjson.LockJSON, reposPath string, version string, profile *lockjson.Profile) {
	repos, err := lockJSON.Repos.FindByPath(reposPath)
	if err != nil {
		repos = nil
	}

	if repos == nil {
		// vim plugin is not found in lock.json
		// -> previous operation is install

		// Add repos to 'repos_path'
		lockJSON.Repos = append(lockJSON.Repos, lockjson.Repos{
			Type:    lockjson.ReposGitType,
			TrxID:   lockJSON.TrxID,
			Path:    reposPath,
			Version: version,
		})
		if !profile.ReposPath.Contains(reposPath) {
			// Add repos to 'profiles[]/repos_path'
			profile.ReposPath = append(profile.ReposPath, reposPath)
		}
	} else {
		// vim plugin is found in lock.json
		// -> previous operation is upgrade
		repos.TrxID = lockJSON.TrxID
		repos.Version = version
	}
}
