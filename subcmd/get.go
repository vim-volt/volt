package subcmd

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/pkg/errors"

	"gopkg.in/src-d/go-git.v4"

	"github.com/vim-volt/volt/config"
	"github.com/vim-volt/volt/fileutil"
	"github.com/vim-volt/volt/gitutil"
	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/pathutil"
	"github.com/vim-volt/volt/plugconf"
	"github.com/vim-volt/volt/subcmd/builder"
	"github.com/vim-volt/volt/transaction"

	multierror "github.com/hashicorp/go-multierror"
)

func init() {
	cmdMap["get"] = &getCmd{}
}

type getCmd struct {
	helped   bool
	lockJSON bool
	upgrade  bool
}

func (cmd *getCmd) ProhibitRootExecution(args []string) bool { return true }

func (cmd *getCmd) FlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Println(`
Usage
  volt get [-help] [-l] [-u] [{repository} ...]

Quick example
  $ volt get tyru/caw.vim     # will install tyru/caw.vim plugin
  $ volt get -u tyru/caw.vim  # will upgrade tyru/caw.vim plugin
  $ volt get -l -u            # will upgrade all plugins in current profile
  $ VOLT_DEBUG=1 volt get tyru/caw.vim  # will output more verbosely

  $ mkdir -p ~/volt/repos/localhost/local/hello/plugin
  $ echo 'command! Hello echom "hello"' >~/volt/repos/localhost/local/hello/plugin/hello.vim
  $ volt get localhost/local/hello     # will add the local repository as a plugin
  $ vim -c Hello                       # will output "hello"

Description
  Install or upgrade given {repository} list, or add local {repository} list as plugins.

  And fetch skeleton plugconf from:
    https://github.com/vim-volt/plugconf-templates
  and install it to:
    $VOLTPATH/plugconf/{repository}.vim

Repository List
  {repository} list (=target to perform installing, upgrading, and so on) is determined as followings:
  * If -l option is specified, all plugins in current profile are used
  * If one or more {repository} arguments are specified, the arguments are used

Action
  The action (install, upgrade, or add only) is determined as follows:
    1. If -u option is specified (upgrade):
      * Upgrade git repositories in {repository} list (static repositories are ignored).
      * Add {repository} list to lock.json (if not found)
    2. Or (install):
      * Fetch {repository} list from remotes
      * Add {repository} list to lock.json (if not found)

Static repository
    Volt can manage a local directory as a repository. It's called "static repository".
    When you have unpublished plugins, or you want to manage ~/.vim/* files as one repository
    (this is useful when you use profile feature, see "volt help profile" for more details),
    static repository is useful.
    All you need is to create a directory in "$VOLTPATH/repos/<repos>".

    When -u was not specified (install) and given repositories exist, volt does not make a request to clone the repositories.
    Therefore, "volt get" tries to fetch repositories but skip it because the directory exists.
    then it adds repositories to lock.json if not found.

      $ mkdir -p ~/volt/repos/localhost/local/hello/plugin
      $ echo 'command! Hello echom "hello"' >~/volt/repos/localhost/local/hello/plugin/hello.vim
      $ volt get localhost/local/hello     # will add the local repository as a plugin
      $ vim -c Hello                       # will output "hello"

Repository path
  {repository}'s format is one of the followings:

  1. {user}/{name}
       This is same as "github.com/{user}/{name}"
  2. {site}/{user}/{name}
  3. https://{site}/{user}/{name}
  4. http://{site}/{user}/{name}

Options`)
		fs.PrintDefaults()
		fmt.Println()
		cmd.helped = true
	}
	fs.BoolVar(&cmd.lockJSON, "l", false, "use all plugins in current profile as targets")
	fs.BoolVar(&cmd.upgrade, "u", false, "upgrade plugins")
	return fs
}

func (cmd *getCmd) Run(args []string) *Error {
	// Parse args
	args, err := cmd.parseArgs(args)
	if err == ErrShowedHelp {
		return nil
	}
	if err != nil {
		return &Error{Code: 10, Msg: "Failed to parse args: " + err.Error()}
	}

	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		return &Error{Code: 11, Msg: "Could not read lock.json: " + err.Error()}
	}

	reposPathList, err := cmd.getReposPathList(args, lockJSON)
	if err != nil {
		return &Error{Code: 12, Msg: "Could not get repos list: " + err.Error()}
	}
	if len(reposPathList) == 0 {
		return &Error{Code: 13, Msg: "No repositories are specified"}
	}

	err = cmd.doGet(reposPathList, lockJSON)
	if err != nil {
		return &Error{Code: 20, Msg: err.Error()}
	}

	return nil
}

func (cmd *getCmd) parseArgs(args []string) ([]string, error) {
	fs := cmd.FlagSet()
	fs.Parse(args)
	if cmd.helped {
		return nil, ErrShowedHelp
	}

	if !cmd.lockJSON && len(fs.Args()) == 0 {
		fs.Usage()
		return nil, errors.New("repository was not given")
	}

	return fs.Args(), nil
}

func (cmd *getCmd) getReposPathList(args []string, lockJSON *lockjson.LockJSON) ([]pathutil.ReposPath, error) {
	var reposPathList []pathutil.ReposPath
	if cmd.lockJSON {
		reposList, err := lockJSON.GetCurrentReposList()
		if err != nil {
			return nil, err
		}
		reposPathList = make([]pathutil.ReposPath, 0, len(reposList))
		for i := range reposList {
			reposPathList = append(reposPathList, reposList[i].Path)
		}
	} else {
		reposPathList = make([]pathutil.ReposPath, 0, len(args))
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

func (cmd *getCmd) doGet(reposPathList []pathutil.ReposPath, lockJSON *lockjson.LockJSON) error {
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

	// Read config.toml
	cfg, err := config.Read()
	if err != nil {
		return errors.Wrap(err, "could not read config.toml")
	}

	done := make(chan getParallelResult, len(reposPathList))
	getCount := 0
	// Invoke installing / upgrading tasks
	for _, reposPath := range reposPathList {
		repos, err := lockJSON.Repos.FindByPath(reposPath)
		if err != nil {
			repos = nil
		}
		if repos == nil || repos.Type == lockjson.ReposGitType {
			go cmd.getParallel(reposPath, repos, cfg, done)
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
		// Update repos[]/version
		if strings.HasPrefix(status, statusPrefixFailed) {
			failed = true
		} else {
			added := cmd.updateReposVersion(lockJSON, r.reposPath, r.reposType, r.hash, profile)
			if added && strings.Contains(status, "already exists") {
				status = fmt.Sprintf(fmtAddedRepos, r.reposPath)
			}
			updatedLockJSON = true
		}
		statusList = append(statusList, status)
	}

	// Sort by status
	sort.Strings(statusList)

	if updatedLockJSON {
		// Write to lock.json
		err = lockJSON.Write()
		if err != nil {
			return errors.Wrap(err, "could not write to lock.json")
		}
	}

	// Build ~/.vim/pack/volt dir
	err = builder.Build(false)
	if err != nil {
		return errors.Wrap(err, "could not build "+pathutil.VimVoltDir())
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
	reposPath pathutil.ReposPath
	status    string
	hash      string
	reposType lockjson.ReposType
	err       error
}

const (
	statusPrefixFailed = "!"
	// Failed
	fmtInstallFailed = "! %s > install failed"
	fmtUpgradeFailed = "! %s > upgrade failed"
	// No change
	fmtNoChange      = "# %s > no change"
	fmtAlreadyExists = "# %s > already exists"
	// Installed
	fmtAddedRepos = "+ %s > added repository to current profile"
	fmtInstalled  = "+ %s > installed"
	// Upgraded
	fmtRevUpdate = "* %s > updated lock.json revision (%s..%s)"
	fmtUpgraded  = "* %s > upgraded (%s..%s)"
	fmtFetched   = "* %s > fetched objects (worktree is not updated)"
)

// This function is executed in goroutine of each plugin.
// 1. install plugin if it does not exist
// 2. install plugconf if it does not exist and createPlugconf=true
func (cmd *getCmd) getParallel(reposPath pathutil.ReposPath, repos *lockjson.Repos, cfg *config.Config, done chan<- getParallelResult) {
	pluginDone := make(chan getParallelResult)
	go cmd.installPlugin(reposPath, repos, cfg, pluginDone)
	pluginResult := <-pluginDone
	if pluginResult.err != nil || !*cfg.Get.CreateSkeletonPlugconf {
		done <- pluginResult
		return
	}
	plugconfDone := make(chan getParallelResult)
	go cmd.installPlugconf(reposPath, &pluginResult, plugconfDone)
	done <- (<-plugconfDone)
}

func (cmd *getCmd) installPlugin(reposPath pathutil.ReposPath, repos *lockjson.Repos, cfg *config.Config, done chan<- getParallelResult) {
	// true:upgrade, false:install
	fullReposPath := reposPath.FullPath()
	doUpgrade := cmd.upgrade && pathutil.Exists(fullReposPath)
	doInstall := !pathutil.Exists(fullReposPath)

	var fromHash string
	var err error
	if doUpgrade {
		// Get HEAD hash string
		fromHash, err = gitutil.GetHEAD(reposPath)
		if err != nil {
			result := errors.Wrap(err, "failed to get HEAD commit hash")
			done <- getParallelResult{
				reposPath: reposPath,
				status:    fmt.Sprintf(fmtInstallFailed, reposPath),
				err:       result,
			}
			return
		}
	}

	var status string
	var upgraded bool
	var checkRevision bool

	if doUpgrade {
		// when cmd.upgrade is true, repos must not be nil.
		if repos == nil {
			done <- getParallelResult{
				reposPath: reposPath,
				status:    fmt.Sprintf(fmtUpgradeFailed, reposPath),
				err:       errors.New("failed to upgrade plugin: -u was specified but repos == nil"),
			}
			return
		}
		// Upgrade plugin
		logger.Debug("Upgrading " + reposPath + " ...")
		err := cmd.upgradePlugin(reposPath, cfg)
		if err != git.NoErrAlreadyUpToDate && err != nil {
			result := errors.Wrap(err, "failed to upgrade plugin")
			done <- getParallelResult{
				reposPath: reposPath,
				status:    fmt.Sprintf(fmtUpgradeFailed, reposPath),
				err:       result,
			}
			return
		}
		if err == git.NoErrAlreadyUpToDate {
			status = fmt.Sprintf(fmtNoChange, reposPath)
		} else {
			upgraded = true
		}
	} else if doInstall {
		// Install plugin
		logger.Debug("Installing " + reposPath + " ...")
		err := cmd.clonePlugin(reposPath, cfg)
		if err != nil {
			result := errors.Wrap(err, "failed to install plugin")
			logger.Debug("Rollbacking " + fullReposPath + " ...")
			err = cmd.removeDir(fullReposPath)
			if err != nil {
				result = multierror.Append(result, err)
			}
			done <- getParallelResult{
				reposPath: reposPath,
				status:    fmt.Sprintf(fmtInstallFailed, reposPath),
				err:       result,
			}
			return
		}
		status = fmt.Sprintf(fmtInstalled, reposPath)
	} else {
		status = fmt.Sprintf(fmtAlreadyExists, reposPath)
		checkRevision = true
	}

	var toHash string
	reposType, err := cmd.detectReposType(fullReposPath)
	if err == nil && reposType == lockjson.ReposGitType {
		// Get HEAD hash string
		toHash, err = gitutil.GetHEAD(reposPath)
		if err != nil {
			result := errors.Wrap(err, "failed to get HEAD commit hash")
			if doInstall {
				logger.Debug("Rollbacking " + fullReposPath + " ...")
				err = cmd.removeDir(fullReposPath)
				if err != nil {
					result = multierror.Append(result, err)
				}
			}
			done <- getParallelResult{
				reposPath: reposPath,
				status:    fmt.Sprintf(fmtInstallFailed, reposPath),
				err:       result,
			}
			return
		}
	}

	if upgraded {
		if fromHash != toHash {
			status = fmt.Sprintf(fmtUpgraded, reposPath, fromHash, toHash)
		} else {
			status = fmt.Sprintf(fmtFetched, reposPath)
		}
	}

	if checkRevision && repos != nil && repos.Version != toHash {
		status = fmt.Sprintf(fmtRevUpdate, reposPath, repos.Version, toHash)
	}

	done <- getParallelResult{
		reposPath: reposPath,
		status:    status,
		reposType: reposType,
		hash:      toHash,
	}
}

func (cmd *getCmd) installPlugconf(reposPath pathutil.ReposPath, pluginResult *getParallelResult, done chan<- getParallelResult) {
	// Install plugconf
	logger.Debug("Installing plugconf " + reposPath + " ...")
	err := cmd.downloadPlugconf(reposPath)
	if err != nil {
		result := errors.Wrap(err, "failed to install plugconf")
		// TODO: Call cmd.removeDir() only when the repos *did not* exist previously
		// and was installed newly.
		// fullReposPath := reposPath.FullPath()
		// logger.Debug("Rollbacking " + fullReposPath + " ...")
		// err = cmd.removeDir(fullReposPath)
		// if err != nil {
		// 	result = multierror.Append(result, err)
		// }
		done <- getParallelResult{
			reposPath: reposPath,
			status:    fmt.Sprintf(fmtInstallFailed, reposPath),
			err:       result,
		}
		return
	}
	done <- *pluginResult
}

func (*getCmd) detectReposType(fullpath string) (lockjson.ReposType, error) {
	if pathutil.Exists(filepath.Join(fullpath, ".git")) {
		if _, err := git.PlainOpen(fullpath); err != nil {
			return "", err
		}
		return lockjson.ReposGitType, nil
	}
	return lockjson.ReposStaticType, nil
}

func (*getCmd) removeDir(fullReposPath string) error {
	if pathutil.Exists(fullReposPath) {
		err := os.RemoveAll(fullReposPath)
		if err != nil {
			return errors.Errorf("rollback failed: cannot remove '%s'", fullReposPath)
		}
		// Remove parent directories
		fileutil.RemoveDirs(filepath.Dir(fullReposPath))
	}
	return nil
}

func (cmd *getCmd) upgradePlugin(reposPath pathutil.ReposPath, cfg *config.Config) error {
	fullpath := reposPath.FullPath()

	repos, err := git.PlainOpen(fullpath)
	if err != nil {
		return err
	}

	reposCfg, err := repos.Config()
	if err != nil {
		return err
	}

	remote, err := gitutil.GetUpstreamRemote(repos)
	if err != nil {
		return err
	}

	if reposCfg.Core.IsBare {
		return cmd.gitFetch(repos, fullpath, remote, cfg)
	}
	return cmd.gitPull(repos, fullpath, remote, cfg)
}

var errRepoExists = errors.New("repository exists")

func (cmd *getCmd) clonePlugin(reposPath pathutil.ReposPath, cfg *config.Config) error {
	fullpath := reposPath.FullPath()
	if pathutil.Exists(fullpath) {
		return errRepoExists
	}

	err := os.MkdirAll(filepath.Dir(fullpath), 0755)
	if err != nil {
		return err
	}

	// Clone repository to $VOLTPATH/repos/{site}/{user}/{name}
	return cmd.gitClone(reposPath.CloneURL(), fullpath, cfg)
}

func (cmd *getCmd) downloadPlugconf(reposPath pathutil.ReposPath) error {
	path := reposPath.Plugconf()
	if pathutil.Exists(path) {
		logger.Debugf("plugconf '%s' exists... skip", path)
		return nil
	}

	// If non-nil error returned from FetchPlugconfTemplate(),
	// create skeleton plugconf file
	tmpl, err := plugconf.FetchPlugconfTemplate(reposPath)
	if err != nil {
		logger.Debug(err.Error())
		// empty tmpl is returned when err != nil
	}
	content, merr := tmpl.Generate(path)
	if merr.ErrorOrNil() != nil {
		return errors.Errorf("parse error in fetched plugconf %s: %s", reposPath, merr.Error())
	}
	os.MkdirAll(filepath.Dir(path), 0755)
	err = ioutil.WriteFile(path, content, 0644)
	if err != nil {
		return err
	}
	return nil
}

// * Add repos to 'repos' if not found
// * Add repos to 'profiles[]/repos_path' if not found
func (*getCmd) updateReposVersion(lockJSON *lockjson.LockJSON, reposPath pathutil.ReposPath, reposType lockjson.ReposType, version string, profile *lockjson.Profile) bool {
	repos, err := lockJSON.Repos.FindByPath(reposPath)
	if err != nil {
		repos = nil
	}

	added := false

	if repos == nil {
		// repos is not found in lock.json
		// -> previous operation is install
		repos = &lockjson.Repos{
			Type:    reposType,
			Path:    reposPath,
			Version: version,
		}
		// Add repos to 'repos'
		lockJSON.Repos = append(lockJSON.Repos, *repos)
		sort.SliceStable(lockJSON.Repos, func(i, j int) bool {
			return strings.ToLower(lockJSON.Repos[i].Path.FullPath()) < strings.ToLower(lockJSON.Repos[j].Path.FullPath())
		})
		added = true
	} else {
		// repos is found in lock.json
		// -> previous operation is upgrade
		repos.Version = version
	}

	if !profile.ReposPath.Contains(reposPath) {
		// Add repos to 'profiles[]/repos_path'
		profile.ReposPath = append(profile.ReposPath, reposPath)
		sort.SliceStable(profile.ReposPath, func(i, j int) bool {
			return strings.ToLower(profile.ReposPath[i].FullPath()) < strings.ToLower(profile.ReposPath[j].FullPath())
		})
		added = true
	}

	return added
}

func (cmd *getCmd) gitFetch(r *git.Repository, workDir string, remote string, cfg *config.Config) error {
	err := r.Fetch(&git.FetchOptions{
		RemoteName: remote,
	})
	if err == nil || err == git.NoErrAlreadyUpToDate {
		return err
	}

	// When fallback_git_cmd is true and git command is installed,
	// try to invoke git-fetch command
	if !*cfg.Get.FallbackGitCmd || !cmd.hasGitCmd() {
		return err
	}
	logger.Warnf("failed to fetch, try to execute \"git fetch %s\" instead...: %s", remote, err.Error())

	before, err := gitutil.GetHEADRepository(r)
	fetch := exec.Command("git", "fetch", remote)
	fetch.Dir = workDir
	err = fetch.Run()
	if err != nil {
		return err
	}
	if changed, err := cmd.getWorktreeChanges(r, before); err != nil {
		return err
	} else if !changed {
		return git.NoErrAlreadyUpToDate
	}
	return nil
}

func (cmd *getCmd) gitPull(r *git.Repository, workDir string, remote string, cfg *config.Config) error {
	wt, err := r.Worktree()
	if err != nil {
		return err
	}
	err = wt.Pull(&git.PullOptions{
		RemoteName: remote,
		// TODO: Temporarily recursive clone is disabled, because go-git does
		// not support relative submodule url in .gitmodules and it causes an
		// error
		RecurseSubmodules: 0,
	})
	if err == nil || err == git.NoErrAlreadyUpToDate {
		return err
	}

	// When fallback_git_cmd is true and git command is installed,
	// try to invoke git-pull command
	if !*cfg.Get.FallbackGitCmd || !cmd.hasGitCmd() {
		return err
	}
	logger.Warnf("failed to pull, try to execute \"git pull\" instead...: %s", err.Error())

	before, err := gitutil.GetHEADRepository(r)
	pull := exec.Command("git", "pull")
	pull.Dir = workDir
	err = pull.Run()
	if err != nil {
		return err
	}
	if changed, err := cmd.getWorktreeChanges(r, before); err != nil {
		return err
	} else if !changed {
		return git.NoErrAlreadyUpToDate
	}
	return nil
}

func (cmd *getCmd) getWorktreeChanges(r *git.Repository, before string) (bool, error) {
	after, err := gitutil.GetHEADRepository(r)
	if err != nil {
		return false, err
	}
	return before != after, nil
}

func (cmd *getCmd) gitClone(cloneURL, dstDir string, cfg *config.Config) error {
	isBare := false
	r, err := git.PlainClone(dstDir, isBare, &git.CloneOptions{
		URL: cloneURL,
		// TODO: Temporarily recursive clone is disabled, because go-git does
		// not support relative submodule url in .gitmodules and it causes an
		// error
		RecurseSubmodules: 0,
	})
	if err != nil {
		// When fallback_git_cmd is true and git command is installed,
		// try to invoke git-clone command
		if !*cfg.Get.FallbackGitCmd || !cmd.hasGitCmd() {
			return err
		}
		logger.Warnf("failed to clone, try to execute \"git clone --recursive %s %s\" instead...: %s", cloneURL, dstDir, err.Error())
		err = os.RemoveAll(dstDir)
		if err != nil {
			return err
		}
		out, err := exec.Command("git", "clone", "--recursive", cloneURL, dstDir).CombinedOutput()
		if err != nil {
			return errors.Errorf("\"git clone --recursive %s %s\" failed, out=%s: %s", cloneURL, dstDir, string(out), err.Error())
		}
	}

	return gitutil.SetUpstreamRemote(r, "origin")
}

func (cmd *getCmd) hasGitCmd() bool {
	exeName := "git"
	if runtime.GOOS == "windows" {
		exeName = "git.exe"
	}
	_, err := exec.LookPath(exeName)
	return err == nil
}
