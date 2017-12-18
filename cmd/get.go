package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/vim-volt/volt/fileutil"
	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/pathutil"
	"github.com/vim-volt/volt/plugconf"
	"github.com/vim-volt/volt/transaction"

	multierror "github.com/hashicorp/go-multierror"
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
	profile, err := lockJSON.Profiles.FindByName(lockJSON.CurrentProfileName)
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
	failed := false
	statusList := make([]string, 0, getCount)
	var updatedLockJSON bool
	for i := 0; i < getCount; i++ {
		r := <-done
		status := cmd.formatStatus(&r)
		// Update repos[]/trx_id, repos[]/version
		if strings.HasPrefix(status, statusPrefixInstalled) ||
			strings.HasPrefix(status, statusPrefixUpgraded) ||
			strings.HasPrefix(status, statusPrefixNoChange) {
			addedProfile := cmd.updateReposVersion(lockJSON, r.reposPath, r.hash, profile)
			if strings.HasPrefix(status, statusPrefixNoChange) && addedProfile {
				status += " > added to current profile"
			}
			updatedLockJSON = true
		}
		if strings.HasPrefix(status, statusPrefixFailed) {
			failed = true
		}
		statusList = append(statusList, status)
	}

	// Sort by status
	sort.Strings(statusList)

	if updatedLockJSON {
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

	// Show results
	for i := range statusList {
		fmt.Println(statusList[i])
	}
	if failed {
		return errors.New("failed to install some plugins")
	}
	return nil
}

func (*getCmd) formatStatus(r *getParallelResult) string {
	if r.err == nil {
		return r.status
	}
	var errs []error
	if merr, ok := r.err.(*multierror.Error); ok {
		errs = merr.Errors
	} else {
		errs = []error{r.err}
	}
	buf := make([]byte, 0, 4*1024)
	buf = append(buf, r.status...)
	for _, err := range errs {
		buf = append(buf, "\n  * "...)
		buf = append(buf, err.Error()...)
	}
	return string(buf)
}

type getParallelResult struct {
	reposPath string
	status    string
	hash      string
	err       error
}

const (
	statusPrefixFailed    = "!"
	statusPrefixNoChange  = "#"
	statusPrefixInstalled = "+"
	statusPrefixUpgraded  = "*"
)

// This function is executed in goroutine of each plugin
func (cmd *getCmd) getParallel(reposPath string, repos *lockjson.Repos, flags *getFlagsType, done chan getParallelResult) {
	const fmtInstallFailed = "%s %s > install failed > %s"
	const fmtUpgradeFailed = "%s %s > upgrade failed > %s"
	const fmtNoChange = "%s %s > no change"
	const fmtAlreadyExists = "%s %s > already exists"
	const fmtInstalled = "%s %s > installed"
	const fmtUpgraded = "%s %s > upgraded (%s..%s)"

	// true:upgrade, false:install
	fullReposPath := pathutil.FullReposPathOf(reposPath)
	doUpgrade := flags.upgrade && pathutil.Exists(fullReposPath)

	var fromHash string
	var err error
	if doUpgrade {
		// Get HEAD hash string
		fromHash, err = getReposHEAD(reposPath)
		if err != nil {
			result := errors.New("failed to get HEAD commit hash: " + err.Error())
			if flags.verbose {
				logger.Info("Rollbacking " + fullReposPath + " ...")
			} else {
				logger.Debug("Rollbacking " + fullReposPath + " ...")
			}
			err = cmd.rollbackRepos(fullReposPath)
			if err != nil {
				result = multierror.Append(result, err)
			}
			done <- getParallelResult{
				reposPath: reposPath,
				status:    fmt.Sprintf(fmtInstallFailed, statusPrefixFailed, reposPath, result.Error()),
				err:       result,
			}
			return
		}
	}

	var status string
	var upgraded bool

	if doUpgrade {
		// when flags.upgrade is true, repos must not be nil.
		if repos == nil {
			msg := "-u was specified but repos == nil"
			done <- getParallelResult{
				reposPath: reposPath,
				status:    fmt.Sprintf(fmtUpgradeFailed, statusPrefixFailed, reposPath, msg),
				err:       errors.New("failed to upgrade plugin: " + msg),
			}
		}
		// Upgrade plugin
		if flags.verbose {
			logger.Info("Upgrading " + reposPath + " ...")
		} else {
			logger.Debug("Upgrading " + reposPath + " ...")
		}
		err := cmd.upgradePlugin(reposPath, flags)
		if err != git.NoErrAlreadyUpToDate && err != nil {
			result := errors.New("failed to upgrade plugin: " + err.Error())
			if flags.verbose {
				logger.Info("Rollbacking " + fullReposPath + " ...")
			} else {
				logger.Debug("Rollbacking " + fullReposPath + " ...")
			}
			err = cmd.rollbackRepos(fullReposPath)
			if err != nil {
				result = multierror.Append(result, err)
			}
			done <- getParallelResult{
				reposPath: reposPath,
				status:    fmt.Sprintf(fmtUpgradeFailed, statusPrefixFailed, reposPath, err.Error()),
				err:       result,
			}
			return
		}
		if err == git.NoErrAlreadyUpToDate {
			status = fmt.Sprintf(fmtNoChange, statusPrefixNoChange, reposPath)
		} else {
			upgraded = true
		}
	} else {
		// Install plugin
		if flags.verbose {
			logger.Info("Installing " + reposPath + " ...")
		} else {
			logger.Debug("Installing " + reposPath + " ...")
		}
		err := cmd.installPlugin(reposPath, flags)
		// if err == errRepoExists, silently skip
		if err != nil && err != errRepoExists {
			result := errors.New("failed to install plugin: " + err.Error())
			if flags.verbose {
				logger.Info("Rollbacking " + fullReposPath + " ...")
			} else {
				logger.Debug("Rollbacking " + fullReposPath + " ...")
			}
			err = cmd.rollbackRepos(fullReposPath)
			if err != nil {
				result = multierror.Append(result, err)
			}
			done <- getParallelResult{
				reposPath: reposPath,
				status:    fmt.Sprintf(fmtInstallFailed, statusPrefixFailed, reposPath, result.Error()),
				err:       result,
			}
			return
		}
		if err == errRepoExists {
			status = fmt.Sprintf(fmtAlreadyExists, statusPrefixNoChange, reposPath)
		} else {
			// Install plugconf
			if flags.verbose {
				logger.Info("Installing plugconf " + reposPath + " ...")
			} else {
				logger.Debug("Installing plugconf " + reposPath + " ...")
			}
			err = cmd.installPlugconf(reposPath)
			if err != nil {
				result := errors.New("failed to install plugconf: " + err.Error())
				if flags.verbose {
					logger.Info("Rollbacking " + fullReposPath + " ...")
				} else {
					logger.Debug("Rollbacking " + fullReposPath + " ...")
				}
				err = cmd.rollbackRepos(fullReposPath)
				if err != nil {
					result = multierror.Append(result, err)
				}
				done <- getParallelResult{
					reposPath: reposPath,
					status:    fmt.Sprintf(fmtInstallFailed, statusPrefixFailed, reposPath, result.Error()),
					err:       result,
				}
				return
			}
			status = fmt.Sprintf(fmtInstalled, statusPrefixInstalled, reposPath)
		}
	}

	// Get HEAD hash string
	var toHash string
	toHash, err = getReposHEAD(reposPath)
	if err != nil {
		result := errors.New("failed to get HEAD commit hash: " + err.Error())
		if flags.verbose {
			logger.Info("Rollbacking " + fullReposPath + " ...")
		} else {
			logger.Debug("Rollbacking " + fullReposPath + " ...")
		}
		err = cmd.rollbackRepos(fullReposPath)
		if err != nil {
			result = multierror.Append(result, err)
		}
		done <- getParallelResult{
			reposPath: reposPath,
			status:    fmt.Sprintf(fmtInstallFailed, statusPrefixFailed, reposPath, result.Error()),
			err:       result,
		}
		return
	}

	// Show old and new revisions: "upgraded ({from}..{to})".
	if upgraded {
		status = fmt.Sprintf(fmtUpgraded, statusPrefixUpgraded, reposPath, fromHash, toHash)
	}

	done <- getParallelResult{
		reposPath: reposPath,
		status:    status,
		hash:      toHash,
	}
}

func (*getCmd) rollbackRepos(fullReposPath string) error {
	if pathutil.Exists(fullReposPath) {
		err := os.RemoveAll(fullReposPath)
		if err != nil {
			return fmt.Errorf("rollback failed: cannot remove '%s'", fullReposPath)
		}
		// Remove parent directories
		fileutil.RemoveDirs(filepath.Dir(fullReposPath))
	}
	return nil
}

func (cmd *getCmd) upgradePlugin(reposPath string, flags *getFlagsType) error {
	fullpath := pathutil.FullReposPathOf(reposPath)

	var progress sideband.Progress = nil
	// if flags.verbose {
	// 	progress = os.Stdout
	// }

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

var errRepoExists = errors.New("repository exists")

func (cmd *getCmd) installPlugin(reposPath string, flags *getFlagsType) error {
	fullpath := pathutil.FullReposPathOf(reposPath)
	if pathutil.Exists(fullpath) {
		return errRepoExists
	}

	var progress sideband.Progress = nil
	// if flags.verbose {
	// 	progress = os.Stdout
	// }

	err := os.MkdirAll(filepath.Dir(fullpath), 0755)
	if err != nil {
		return err
	}

	// Clone repository to $VOLTPATH/repos/{site}/{user}/{name}
	isBare := false
	r, err := git.PlainClone(fullpath, isBare, &git.CloneOptions{
		URL:      pathutil.CloneURLOf(reposPath),
		Progress: progress,
	})
	if err != nil {
		return err
	}

	return cmd.setUpstreamBranch(r)
}

func (cmd *getCmd) setUpstreamBranch(r *git.Repository) error {
	cfg, err := r.Config()
	if err != nil {
		return err
	}

	head, err := r.Head()
	if err != nil {
		return err
	}

	refBranch := head.Name().String()
	branch := refHeadsRx.FindStringSubmatch(refBranch)
	if len(branch) == 0 {
		return errors.New("HEAD is not matched to refs/heads/...: " + refBranch)
	}

	sec := cfg.Raw.Section("branch")
	subsec := sec.Subsection(branch[1])
	subsec.AddOption("remote", "origin")
	subsec.AddOption("merge", refBranch)

	return r.Storer.SetConfig(cfg)
}

func (cmd *getCmd) installPlugconf(reposPath string) error {
	filename := pathutil.PlugconfOf(reposPath)
	if pathutil.Exists(filename) {
		logger.Debugf("plugconf '%s' exists... skip", filename)
		return nil
	}

	// If non-nil error returned from FetchPlugconf(),
	// create skeleton plugconf file
	tmpl, err := plugconf.FetchPlugconf(reposPath)
	if err != nil {
		logger.Debug(err.Error())
	}
	content, err := plugconf.GenPlugconfByTemplate(tmpl, filename)
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

// * Add repos to 'repos' if not found
// * Add repos to 'profiles[]/repos_path' if not found
func (*getCmd) updateReposVersion(lockJSON *lockjson.LockJSON, reposPath string, version string, profile *lockjson.Profile) bool {
	repos, err := lockJSON.Repos.FindByPath(reposPath)
	if err != nil {
		repos = nil
	}

	if repos == nil {
		// repos is not found in lock.json
		// -> previous operation is install
		repos = &lockjson.Repos{
			Type:    lockjson.ReposGitType,
			TrxID:   lockJSON.TrxID,
			Path:    reposPath,
			Version: version,
		}
		// Add repos to 'repos'
		lockJSON.Repos = append(lockJSON.Repos, *repos)
	} else {
		// repos is found in lock.json
		// -> previous operation is upgrade
		repos.TrxID = lockJSON.TrxID
		repos.Version = version
	}

	addedProfile := false
	if !profile.ReposPath.Contains(reposPath) {
		// Add repos to 'profiles[]/repos_path'
		profile.ReposPath = append(profile.ReposPath, reposPath)
		addedProfile = true
	}
	return addedProfile
}
