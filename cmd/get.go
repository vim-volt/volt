package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/vim-volt/go-volt/lockjson"
	"github.com/vim-volt/go-volt/pathutil"
	"github.com/vim-volt/go-volt/transaction"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing/format/gitignore"
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
		fmt.Println("[ERROR] Failed to parse args: " + err.Error())
		return 10
	}

	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		fmt.Println("[ERROR] Could not read lock.json: " + err.Error())
		return 11
	}

	reposPathList, err := cmd.getReposPathList(flags, args, lockJSON)
	if err != nil {
		fmt.Println("[ERROR] Could not get repos list: " + err.Error())
		return 12
	}

	// Parse global gitignore file
	ps, err := cmd.getGlobalGitignore()
	if err != nil {
		fmt.Println("[WARN] Could not get global gitignore config: " + err.Error())
		ps = nil
	}

	// Check if any repositories are dirty
	for _, reposPath := range reposPathList {
		fullpath := pathutil.FullReposPathOf(reposPath)
		if cmd.pathExists(fullpath) && cmd.isDirtyWorktree(fullpath, ps) {
			fmt.Println("[ERROR] Repository has dirty worktree: " + fullpath)
			return 14
		}
	}

	// Begin transaction
	err = transaction.Create()
	if err != nil {
		fmt.Println("[ERROR] Failed to begin transaction: " + err.Error())
		return 15
	}
	defer transaction.Remove()
	lockJSON.TrxID++

	var updatedLockJSON bool
	var upgradedList []string
	for _, reposPath := range reposPathList {
		upgrade := flags.upgrade && cmd.pathExists(pathutil.FullReposPathOf(reposPath))

		// Install / Upgrade plugin
		err = cmd.installPlugin(reposPath, flags)
		if err != nil {
			fmt.Println("[ERROR] Failed to install / upgrade plugins: " + err.Error())
			return 16
		}

		// Fetch plugconf
		fmt.Println("[INFO] Installing plugconf " + reposPath + " ...")

		err = cmd.installPlugConf(reposPath + ".vim")
		err2 := cmd.installPlugConf(reposPath + ".json")
		if err != nil && err2 != nil {
			fmt.Println("[INFO] Not found plugconf")
		} else {
			list := make([]string, 0, 2)
			if err == nil {
				list = append(list, "vim")
			}
			if err == nil {
				list = append(list, "json")
			}
			fmt.Println("[INFO] Found plugconf (" + strings.Join(list, ",") + ")")
		}

		// Get HEAD hash string
		hash, err := cmd.getHEADHashString(reposPath)
		if err != nil {
			fmt.Println("[ERROR] Failed to get HEAD commit hash: " + err.Error())
			continue
		}
		// Update repos[]/trx_id, repos[]/version
		cmd.updateReposVersion(lockJSON, reposPath, hash)
		updatedLockJSON = true
		// Collect upgraded repos path
		if upgrade {
			upgradedList = append(upgradedList, reposPath)
		}
	}

	if updatedLockJSON {
		err = lockjson.Write(lockJSON)
		if err != nil {
			fmt.Println("[ERROR] Could not write to lock.json: " + err.Error())
			return 16
		}
	}

	// Show upgraded plugins
	if len(upgradedList) > 0 {
		fmt.Println("[WARN] Reloading upgraded plugin is not supported.")
		fmt.Println("[WARN] Please restart your Vim to reload the following plugins:")
		for _, reposPath := range upgradedList {
			fmt.Println("[WARN]   " + reposPath)
		}
	}

	return 0
}

func (getCmd) parseArgs(args []string) ([]string, *getFlags, error) {
	var flags getFlags
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Println(`
Usage
  volt get [-help] [-l] [-u] [-v] [{repository} ...]

Description
  Install / Upgrade vim plugin, and system plugconf files from
  https://github.com/vim-volt/plugconf-templates

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
	return fs.Args(), &flags, nil
}

func (getCmd) getReposPathList(flags *getFlags, args []string, lockJSON *lockjson.LockJSON) ([]string, error) {
	reposPathList := make([]string, 0, 32)
	if flags.lockJSON {
		for _, repos := range lockJSON.Repos {
			reposPathList = append(reposPathList, repos.Path)
		}
	} else {
		for _, arg := range args {
			reposPath, err := pathutil.NormalizeRepository(arg)
			if err != nil {
				return nil, err
			}
			reposPathList = append(reposPathList, reposPath)
		}
	}
	return reposPathList, nil
}

func (cmd getCmd) getGlobalGitignore() ([]gitignore.Pattern, error) {
	cfg, err := cmd.parseGitConfig()
	if err != nil {
		return nil, errors.New("could not read ~/.gitconfig: " + err.Error())
	}

	excludesfile := cmd.getExcludesFile(cfg)
	if excludesfile == "" {
		return nil, errors.New("could not get core.excludesfile from ~/.gitconfig")
	}

	ps, err := cmd.parseExcludesFile(excludesfile)
	if err != nil {
		return nil, errors.New("could not parse core.excludesfile: " + excludesfile + ": " + err.Error())
	}
	return ps, nil
}

func (getCmd) parseGitConfig() (*config.Config, error) {
	cfg := config.NewConfig()

	u, err := user.Current()
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadFile(filepath.Join(u.HomeDir, ".gitconfig"))
	if err != nil {
		return nil, err
	}

	if err := cfg.Unmarshal(b); err != nil {
		return nil, err
	}

	return cfg, err
}

func (getCmd) getExcludesFile(cfg *config.Config) string {
	for _, sec := range cfg.Raw.Sections {
		if sec.Name == "core" {
			for _, opt := range sec.Options {
				if opt.Key == "excludesfile" {
					return opt.Value
				}
			}
		}
	}
	return ""
}

func (cmd getCmd) parseExcludesFile(excludesfile string) ([]gitignore.Pattern, error) {
	excludesfile, err := cmd.expandTilde(excludesfile)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(excludesfile)
	if err != nil {
		return nil, err
	}

	var ps []gitignore.Pattern
	for _, s := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(s, "#") && len(strings.TrimSpace(s)) > 0 {
			ps = append(ps, gitignore.ParsePattern(s, nil))
		}
	}

	return ps, nil
}

// "~/.gitignore" -> "/home/tyru/.gitignore"
func (getCmd) expandTilde(path string) (string, error) {
	var paths []string
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	for _, p := range strings.Split(path, string(filepath.Separator)) {
		if p == "~" {
			paths = append(paths, u.HomeDir)
		} else {
			paths = append(paths, p)
		}
	}
	return filepath.Join(paths...), nil
}

func (getCmd) pathExists(fullpath string) bool {
	_, err := os.Stat(fullpath)
	return !os.IsNotExist(err)
}

func (getCmd) isDirtyWorktree(fullpath string, ps []gitignore.Pattern) bool {
	repos, err := git.PlainOpen(fullpath)
	if err != nil {
		return true
	}
	wt, err := repos.Worktree()
	if err != nil {
		return true
	}
	st, err := wt.Status()
	if err != nil {
		return true
	}

	return !st.IsClean()

	// TODO: Instead of using IsClean(), check each file with ignored patterns
	//
	// paths := make([]string, 0, len(st))
	// for path := range st {
	// 	paths = append(paths, path)
	// }
	//
	// m := gitignore.NewMatcher(ps)
	// return !m.Match(paths, false)
}

func (cmd getCmd) installPlugin(reposPath string, flags *getFlags) error {
	fullpath := pathutil.FullReposPathOf(reposPath)
	if !flags.upgrade && cmd.pathExists(fullpath) {
		return errors.New("repository exists")
	}

	if flags.upgrade {
		fmt.Println("[INFO] Upgrading " + reposPath + " ...")
	} else {
		fmt.Println("[INFO] Installing " + reposPath + " ...")
	}

	// Get existing temporary directory path
	tempPath, err := cmd.getTempPath()
	if err != nil {
		return err
	}

	var progress sideband.Progress = nil
	if flags.verbose {
		progress = os.Stdout
	}

	// git clone to temporary directory
	tempGitRepos, err := git.PlainClone(tempPath, false, &git.CloneOptions{
		URL:      pathutil.CloneURLOf(reposPath),
		Progress: progress,
	})
	if err != nil {
		return err
	}

	// If !flags.upgrade or HEAD was changed (= the plugin is outdated) ...
	if !flags.upgrade || cmd.headWasChanged(reposPath, tempGitRepos) {
		// Remove existing repository
		if cmd.pathExists(fullpath) {
			err = os.RemoveAll(fullpath)
			if err != nil {
				return err
			}
		}

		// Move repository to $VOLTPATH/repos/{site}/{user}/{name}
		err = os.MkdirAll(filepath.Dir(fullpath), 0755)
		if err != nil {
			return err
		}
		err = os.Rename(tempPath, fullpath)
		if err != nil {
			return err
		}
	}

	return nil
}

func (getCmd) getTempPath() (string, error) {
	err := os.MkdirAll(pathutil.TempPath(), 0755)
	if err != nil {
		return "", err
	}

	tempPath, err := ioutil.TempDir(pathutil.TempPath(), "volt-")
	if err != nil {
		return "", err
	}

	err = os.MkdirAll(tempPath, 0755)
	if err != nil {
		return "", err
	}
	return tempPath, nil
}

func (cmd getCmd) headWasChanged(reposPath string, tempGitRepos *git.Repository) bool {
	tempHead, err := tempGitRepos.Head()
	if err != nil {
		return false
	}
	hash, err := cmd.getHEADHashString(reposPath)
	if err != nil {
		return false
	}
	return tempHead.Hash().String() != hash
}

func (getCmd) updateReposVersion(lockJSON *lockjson.LockJSON, reposPath string, version string) {
	var r *lockjson.Repos
	for i := range lockJSON.Repos {
		if lockJSON.Repos[i].Path == reposPath {
			r = &lockJSON.Repos[i]
			break
		}
	}

	if r == nil {
		// vim plugin is not found in lock.json
		// -> previous operation is install
		lockJSON.Repos = append(lockJSON.Repos, lockjson.Repos{
			TrxID:   lockJSON.TrxID,
			Path:    reposPath,
			Version: version,
			Active:  true,
		})
	} else {
		// vim plugin is found in lock.json
		// -> previous operation is upgrade
		r.TrxID = lockJSON.TrxID
		r.Version = version
	}
}

func (getCmd) getHEADHashString(reposPath string) (string, error) {
	repos, err := git.PlainOpen(pathutil.FullReposPathOf(reposPath))
	if err != nil {
		return "", err
	}
	head, err := repos.Head()
	if err != nil {
		return "", err
	}
	return head.Hash().String(), nil
}

func (getCmd) installPlugConf(filename string) error {
	url := "https://raw.githubusercontent.com/vim-volt/plugconf-templates/master/templates/" + filename

	res, err := http.Get(url)
	if err != nil {
		return err
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
