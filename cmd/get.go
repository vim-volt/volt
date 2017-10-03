package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/pathutil"
	"github.com/vim-volt/volt/transaction"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/protocol/packp/sideband"
)

type getCmd struct{}

type getFlags struct {
	lockJSON bool
	upgrade  bool
	verbose  bool
}

func Get(args []string) int {
	cmd := getCmd{}

	// Parse args
	args, flags, err := cmd.parseArgs(args)
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

func (*getCmd) parseArgs(args []string) ([]string, *getFlags, error) {
	var flags getFlags
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Println(`
Usage
  volt get [-help] [-l] [-u] [-v] [{repository} ...]

Description
  Install vim plugin from {repository}, or upgrade vim plugin of {repository} list. And fetch system plugconf files from:
    https://github.com/vim-volt/plugconf-templates
  and install it to:
    $VOLTPATH/plugconf/system/{repository}.vim

  {repository}'s format is one of the followings:

  1. {user}/{name}
       This is same as "github.com/{user}/{name}"
  2. {site}/{user}/{name}
  3. https://{site}/{user}/{name}
  4. http://{site}/{user}/{name}

  {repository} list is determined as followings:

  * If -l option and -u option is specified, installed all vim plugins (regardless current profile) are used
  * If {repository} arguments are specified, the specified vim plugins are used

  If both are specified, just error message will be returned.

  If -l and -u options were specified (two options must be used together), upgrade git repositories of installed all vim plugins (static repositories are ignored).

  If -v option was specified, show git-clone(1) output too.

Options`)
		fs.PrintDefaults()
		fmt.Println()
	}
	fs.BoolVar(&flags.lockJSON, "l", false, "from lock.json")
	fs.BoolVar(&flags.upgrade, "u", false, "upgrade installed vim plugin")
	fs.BoolVar(&flags.verbose, "v", false, "show git-clone output")
	fs.Parse(args)

	if !flags.lockJSON && len(fs.Args()) == 0 {
		fs.Usage()
		return nil, nil, errors.New("repository was not given")
	}

	if flags.lockJSON && !flags.upgrade {
		fs.Usage()
		return nil, nil, errors.New("-l must be used with -u")
	}

	return fs.Args(), &flags, nil
}

func (*getCmd) getReposPathList(flags *getFlags, args []string, lockJSON *lockjson.LockJSON) ([]string, error) {
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

func (cmd *getCmd) doGet(reposPathList []string, flags *getFlags, lockJSON *lockjson.LockJSON) error {
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
		return errors.New("failed to begin transaction: " + err.Error())
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
		if r.status != "" {
			statusList = append(statusList, r.status)
		}
		// Update repos[]/trx_id, repos[]/version
		if r.reposPath != "" && r.hash != "" {
			cmd.updateReposVersion(lockJSON, r.reposPath, r.hash, profile)
			updatedLockJSON = true
		}
	}

	// Write to lock.json
	if updatedLockJSON {
		err = lockJSON.Write()
		if err != nil {
			return errors.New("could not write to lock.json: " + err.Error())
		}
	}

	// Rebuild start dir
	err = (&rebuildCmd{}).doRebuild(false)
	if err != nil {
		return errors.New("could not rebuild " + pathutil.VimVoltDir() + ": " + err.Error())
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

func (cmd *getCmd) getParallel(reposPath string, flags *getFlags, done chan getParallelResult) {
	var status string

	if flags.upgrade && pathutil.Exists(pathutil.FullReposPathOf(reposPath)) {
		// Upgrade plugin
		err := cmd.upgradePlugin(reposPath, flags)
		if err != git.NoErrAlreadyUpToDate && err != nil {
			logger.Warn("Failed to upgrade plugin: " + err.Error())

			done <- getParallelResult{
				reposPath: reposPath,
				status:    "! " + reposPath + " : upgrade failed",
			}
			return
		}
		if err == git.NoErrAlreadyUpToDate {
			status = "# " + reposPath + " : no change"
		} else {
			status = "* " + reposPath + " : upgraded"
		}
	} else {
		// Install plugin
		err := cmd.installPlugin(reposPath, flags)
		if err != nil {
			logger.Warn("Failed to install plugin: " + err.Error())
			done <- getParallelResult{
				reposPath: reposPath,
				status:    "! " + reposPath + " : install failed",
			}
			return
		}
		status = "+ " + reposPath + " : installed"

		// Install plugconf
		logger.Info("Installing plugconf " + reposPath + " ...")
		err = cmd.installPlugConf(reposPath + ".vim")
		if err != nil {
			logger.Info("Installing plugconf " + reposPath + " ... not found")
		} else {
			logger.Info("Installing plugconf " + reposPath + " ... found")
		}
	}

	// Get HEAD hash string
	hash, err := cmd.getRemoteHEAD(reposPath)
	if err != nil {
		logger.Error("Failed to get HEAD commit hash: " + err.Error())
		done <- getParallelResult{
			reposPath: reposPath,
			status:    "! " + reposPath + " : install failed",
		}
		return
	}

	done <- getParallelResult{
		reposPath: reposPath,
		status:    status,
		hash:      hash,
	}
}

func (cmd *getCmd) upgradePlugin(reposPath string, flags *getFlags) error {
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

	return repos.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		Progress:   progress,
	})
}

func (cmd *getCmd) installPlugin(reposPath string, flags *getFlags) error {
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
	isBare := true
	_, err = git.PlainClone(fullpath, isBare, &git.CloneOptions{
		URL:      pathutil.CloneURLOf(reposPath),
		Progress: progress,
	})
	return err
}

func (*getCmd) installPlugConf(filename string) error {
	url := "https://raw.githubusercontent.com/vim-volt/plugconf-templates/master/templates/" + filename

	res, err := http.Get(url)
	if err != nil {
		return err
	}
	if res.StatusCode%100 != 2 { // Not 2xx status code
		return errors.New("Returned non-successful status: " + res.Status)
	}
	defer res.Body.Close()

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	fn := pathutil.SystemPlugConfOf(filename)
	dir, _ := filepath.Split(fn)
	os.MkdirAll(dir, 0755)

	err = ioutil.WriteFile(fn, bytes, 0644)
	if err != nil {
		return err
	}
	return nil
}

var refHeadBranchRx = regexp.MustCompile(`^refs/heads/(.+)$`)

func (*getCmd) getRemoteHEAD(reposPath string) (string, error) {
	repos, err := git.PlainOpen(pathutil.FullReposPathOf(reposPath))
	if err != nil {
		return "", err
	}

	head, err := repos.Head()
	if err != nil {
		return "", err
	}

	// e.g. head.Name() = "refs/heads/master"
	match := refHeadBranchRx.FindStringSubmatch(head.Name().String())
	if len(match) == 0 {
		return "", errors.New("could not find branch name from HEAD")
	}

	// Get reference of refs/remotes/origin/{branchName}
	ref, err := repos.Reference(plumbing.ReferenceName("refs/remotes/origin/"+match[1]), true)
	if err != nil {
		return "", err
	}

	return ref.Hash().String(), nil
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
