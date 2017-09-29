package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/vim-volt/go-volt/lockjson"
	"github.com/vim-volt/go-volt/pathutil"
	"github.com/vim-volt/go-volt/transaction"

	"gopkg.in/src-d/go-git.v4"
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

	// Find matching profile
	profile, err := lockJSON.Profiles.FindByName(lockJSON.ActiveProfile)
	if err != nil {
		// this must not be occurred because lockjson.Read()
		// validates if the matching profile exists
		fmt.Println("[ERROR]", err.Error())
		return 15
	}

	// Begin transaction
	err = transaction.Create()
	if err != nil {
		fmt.Println("[ERROR] Failed to begin transaction: " + err.Error())
		return 16
	}
	defer transaction.Remove()
	lockJSON.TrxID++

	var updatedLockJSON bool
	var results []string
	for _, reposPath := range reposPathList {
		if flags.upgrade && cmd.pathExists(pathutil.FullReposPathOf(reposPath)) {
			// Upgrade plugin
			err = cmd.upgradePlugin(reposPath, flags)
			if err != git.NoErrAlreadyUpToDate && err != nil {
				fmt.Println("[WARN] Failed to upgrade plugin: " + err.Error())
				results = append(results, "! "+reposPath+" : upgrade failed")
				continue
			}
			if err == git.NoErrAlreadyUpToDate {
				results = append(results, "# "+reposPath+" : no change")
			} else {
				results = append(results, "* "+reposPath+" : upgraded")
			}
		} else {
			// Install plugin
			err = cmd.installPlugin(reposPath, flags)
			if err != nil {
				fmt.Println("[WARN] Failed to install plugin: " + err.Error())
				results = append(results, "! "+reposPath+" : install failed")
				continue
			}
			results = append(results, "+ "+reposPath+" : installed")

			// Install plugconf
			fmt.Println("[INFO] Installing plugconf " + reposPath + " ...")
			err = cmd.installPlugConf(reposPath + ".vim")
			if err != nil {
				fmt.Println("[INFO] Not found plugconf")
			} else {
				fmt.Println("[INFO] Found plugconf")
			}
		}

		// Get HEAD hash string
		hash, err := cmd.getHEADHashString(reposPath)
		if err != nil {
			fmt.Println("[ERROR] Failed to get HEAD commit hash: " + err.Error())
			continue
		}

		// Update repos[]/trx_id, repos[]/version
		cmd.updateReposVersion(lockJSON, reposPath, hash, profile)
		updatedLockJSON = true
	}

	// Write to lock.json
	if updatedLockJSON {
		err = lockJSON.Write()
		if err != nil {
			fmt.Println("[ERROR] Could not write to lock.json: " + err.Error())
			return 19
		}
	}

	// Rebuild start dir
	err = (&rebuildCmd{}).doRebuild()
	if err != nil {
		fmt.Println("[ERROR] Could not rebuild " + pathutil.VimVoltStartDir() + ": " + err.Error())
		return 20
	}

	// Show results
	if len(results) > 0 {
		fmt.Print("\nDone!\n\n")
		for i := range results {
			fmt.Println(results[i])
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

	if flags.lockJSON && !flags.upgrade {
		fs.Usage()
		return nil, nil, errors.New("-l must be used with -u")
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

func (getCmd) pathExists(fullpath string) bool {
	_, err := os.Stat(fullpath)
	return !os.IsNotExist(err)
}

func (cmd getCmd) upgradePlugin(reposPath string, flags *getFlags) error {
	fullpath := pathutil.FullReposPathOf(reposPath)

	fmt.Println("[INFO] Upgrading " + reposPath + " ...")

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

func (cmd getCmd) installPlugin(reposPath string, flags *getFlags) error {
	fullpath := pathutil.FullReposPathOf(reposPath)
	if cmd.pathExists(fullpath) {
		return errors.New("repository exists")
	}

	fmt.Println("[INFO] Installing " + reposPath + " ...")

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

func (getCmd) installPlugConf(filename string) error {
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

func (getCmd) updateReposVersion(lockJSON *lockjson.LockJSON, reposPath string, version string, profile *lockjson.Profile) {
	repos, err := lockJSON.Repos.FindByPath(reposPath)
	if err != nil {
		repos = nil
	}

	if repos == nil {
		// vim plugin is not found in lock.json
		// -> previous operation is install

		// Add repos to 'repos_path'
		lockJSON.Repos = append(lockJSON.Repos, lockjson.Repos{
			TrxID:   lockJSON.TrxID,
			Path:    reposPath,
			Version: version,
		})
		// Add repos to 'profiles[]/repos_path'
		profile.ReposPath = append(profile.ReposPath, reposPath)
	} else {
		// vim plugin is found in lock.json
		// -> previous operation is upgrade
		repos.TrxID = lockJSON.TrxID
		repos.Version = version
	}
}
