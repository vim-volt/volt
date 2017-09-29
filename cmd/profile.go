package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/vim-volt/go-volt/lockjson"
	"github.com/vim-volt/go-volt/pathutil"
	"github.com/vim-volt/go-volt/transaction"
)

type profileCmd struct {
	showedUsage bool
}

var profileSubCmd = make(map[string]func([]string) error)

func init() {
	cmd := profileCmd{}
	profileSubCmd["get"] = cmd.doGet
	profileSubCmd["set"] = cmd.doSet
	profileSubCmd["show"] = cmd.doShow
	profileSubCmd["list"] = cmd.doList
	profileSubCmd["new"] = cmd.doNew
	profileSubCmd["destroy"] = cmd.doDestroy
	profileSubCmd["add"] = cmd.doAdd
	profileSubCmd["rm"] = cmd.doRm
}

func Profile(args []string) int {
	cmd := profileCmd{}

	// Parse args
	args, err := cmd.parseArgs(args)
	if err != nil {
		fmt.Println(err.Error())
		return 10
	}

	if cmd.showedUsage {
		return 0
	}

	if fn, exists := profileSubCmd[args[0]]; exists {
		err = fn(args[1:])
		if err != nil {
			fmt.Println("[ERROR]", err.Error())
			return 11
		}
	}

	return 0
}

func (cmd *profileCmd) showUsage() {
	cmd.showedUsage = true
	fmt.Println(`
Usage
  profile [get]
    Get current profile name

  profile set {name}
    Set profile name

  profile show {name}
    Show profile info

  profile list
    List all profiles

  profile new {name}
    Create new profile

  profile destroy {name}
    Delete profile

  profile add {name} {repository} [{repository2} ...]
    Add one or more repositories to profile

  profile rm {name} {repository} [{repository2} ...]
    Remove one or more repositories to profile

Description
  Subcommands about profile feature
`)
}

func (cmd *profileCmd) parseArgs(args []string) ([]string, error) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = cmd.showUsage
	fs.Parse(args)

	if len(fs.Args()) == 0 {
		return append([]string{"get"}, fs.Args()...), nil
	}

	subCmd := fs.Args()[0]
	if _, exists := profileSubCmd[subCmd]; !exists {
		return nil, errors.New("unknown subcommand: " + subCmd)
	}
	return fs.Args(), nil
}

func (cmd *profileCmd) doGet(_ []string) error {
	currentProfile, err := cmd.getCurrentProfile()
	if err != nil {
		return err
	}
	fmt.Println(currentProfile)
	return nil
}

func (*profileCmd) getCurrentProfile() (string, error) {
	lockJSON, err := lockjson.Read()
	if err != nil {
		return "", errors.New("failed to read lock.json: " + err.Error())
	}
	return lockJSON.ActiveProfile, nil
}

func (cmd *profileCmd) doSet(args []string) error {
	if len(args) == 0 {
		cmd.showUsage()
		fmt.Println("[ERROR] 'volt profile set' receives profile name.")
		return nil
	}
	profileName := args[0]

	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		return errors.New("failed to read lock.json: " + err.Error())
	}

	// Exit if current active_profile is same as profileName
	if lockJSON.ActiveProfile == profileName {
		fmt.Println("[INFO] Unchanged active profile '" + profileName + "'")
		return nil
	}

	// Begin transaction
	err = transaction.Create()
	if err != nil {
		return err
	}
	defer transaction.Remove()

	// Return error if profiles[]/name does not match profileName
	_, err = lockJSON.Profiles.FindByName(profileName)
	if err != nil {
		return err
	}

	// Set profile name
	lockJSON.ActiveProfile = profileName

	// Write to lock.json
	err = lockJSON.Write()
	if err != nil {
		return err
	}

	fmt.Println("[INFO] Set active profile to '" + profileName + "'")

	// Rebuild start dir
	err = (&rebuildCmd{}).doRebuild()
	if err != nil {
		return errors.New("could not rebuild " + pathutil.VimVoltDir() + ": " + err.Error())
	}

	return nil
}

func (cmd *profileCmd) doShow(args []string) error {
	if len(args) == 0 {
		cmd.showUsage()
		fmt.Println("[ERROR] 'volt profile show' receives profile name.")
		return nil
	}
	profileName := args[0]

	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		return errors.New("failed to read lock.json: " + err.Error())
	}

	// Return error if profiles[]/name does not match profileName
	profile, err := lockJSON.Profiles.FindByName(profileName)
	if err != nil {
		return err
	}

	fmt.Println("name:", profile.Name)
	fmt.Println("load_vimrc:", profile.LoadVimrc)
	fmt.Println("load_gvimrc:", profile.LoadGvimrc)
	fmt.Println("repos_path:")
	for _, reposPath := range profile.ReposPath {
		fmt.Println("  " + reposPath)
	}

	return nil
}

func (cmd *profileCmd) doList(args []string) error {
	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		return errors.New("failed to read lock.json: " + err.Error())
	}

	// List profile names
	for _, profile := range lockJSON.Profiles {
		if profile.Name == lockJSON.ActiveProfile {
			fmt.Println("* " + profile.Name)
		} else {
			fmt.Println("  " + profile.Name)
		}
	}

	return nil
}

func (cmd *profileCmd) doNew(args []string) error {
	if len(args) == 0 {
		cmd.showUsage()
		fmt.Println("[ERROR] 'volt profile new' receives profile name.")
		return nil
	}
	profileName := args[0]

	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		return errors.New("failed to read lock.json: " + err.Error())
	}

	// Return error if profiles[]/name matches profileName
	_, err = lockJSON.Profiles.FindByName(profileName)
	if err == nil {
		return errors.New("profile '" + profileName + "' already exists")
	}

	// Begin transaction
	err = transaction.Create()
	if err != nil {
		return err
	}
	defer transaction.Remove()

	// Add profile
	lockJSON.Profiles = append(lockJSON.Profiles, lockjson.Profile{
		Name:       profileName,
		ReposPath:  make([]string, 0),
		LoadVimrc:  true,
		LoadGvimrc: true,
	})

	// Write to lock.json
	err = lockJSON.Write()
	if err != nil {
		return err
	}

	fmt.Println("[INFO] Created new profile '" + profileName + "'")

	return nil
}

func (cmd *profileCmd) doDestroy(args []string) error {
	if len(args) == 0 {
		cmd.showUsage()
		fmt.Println("[ERROR] 'volt profile destroy' receives profile name.")
		return nil
	}
	profileName := args[0]

	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		return errors.New("failed to read lock.json: " + err.Error())
	}

	// Return error if active_profile matches profileName
	if lockJSON.ActiveProfile == profileName {
		return errors.New("cannot destroy active profile: " + profileName)
	}

	// Return error if profiles[]/name does not match profileName
	index := lockJSON.Profiles.FindIndexByName(profileName)
	if index < 0 {
		return errors.New("profile '" + profileName + "' does not exist")
	}

	// Begin transaction
	err = transaction.Create()
	if err != nil {
		return err
	}
	defer transaction.Remove()

	// Delete the specified profile
	lockJSON.Profiles = append(lockJSON.Profiles[:index], lockJSON.Profiles[index+1:]...)

	// Write to lock.json
	err = lockJSON.Write()
	if err != nil {
		return err
	}

	fmt.Println("[INFO] Deleted profile '" + profileName + "'")

	return nil
}

func (cmd *profileCmd) doAdd(args []string) error {
	// Parse args
	profileName, reposPathList, err := cmd.parseAddArgs("add", args)

	// Read modified profile and write to lock.json
	lockJSON, err := cmd.transactProfile(profileName, func(profile *lockjson.Profile) {
		// Add repositories to profile if the repository does not exist
		for _, reposPath := range reposPathList {
			if profile.ReposPath.Contains(reposPath) {
				fmt.Println("[WARN] repository '" + reposPath + "' already exists")
			} else {
				profile.ReposPath = append(profile.ReposPath, reposPath)
				fmt.Println("[INFO] Added repository '" + reposPath + "'")
			}
		}
	})
	if err != nil {
		return err
	}

	if lockJSON.ActiveProfile == profileName {
		// Rebuild start dir
		err = (&rebuildCmd{}).doRebuild()
		if err != nil {
			return errors.New("could not rebuild " + pathutil.VimVoltDir() + ": " + err.Error())
		}
	}

	return nil
}

func (cmd *profileCmd) doRm(args []string) error {
	// Parse args
	profileName, reposPathList, err := cmd.parseAddArgs("rm", args)

	// Read modified profile and write to lock.json
	lockJSON, err := cmd.transactProfile(profileName, func(profile *lockjson.Profile) {
		// Remove repositories from profile if the repository does not exist
		for _, reposPath := range reposPathList {
			index := profile.ReposPath.IndexOf(reposPath)
			if index >= 0 {
				// Remove profile.ReposPath[index]
				profile.ReposPath = append(profile.ReposPath[:index], profile.ReposPath[index+1:]...)
				fmt.Println("[INFO] Removed repository '" + reposPath + "'")
			} else {
				fmt.Println("[WARN] repository '" + reposPath + "' does not exist")
			}
		}
	})
	if err != nil {
		return err
	}

	if lockJSON.ActiveProfile == profileName {
		// Rebuild start dir
		err = (&rebuildCmd{}).doRebuild()
		if err != nil {
			return errors.New("could not rebuild " + pathutil.VimVoltDir() + ": " + err.Error())
		}
	}

	return nil
}

func (cmd *profileCmd) parseAddArgs(subCmd string, args []string) (string, []string, error) {
	if len(args) == 0 {
		cmd.showUsage()
		fmt.Printf("[ERROR] 'volt profile %s' receives profile name and one or more repositories.\n", subCmd)
		return "", nil, nil
	}

	profileName := args[0]
	reposPathList := make([]string, 0, len(args)-1)
	for _, arg := range args[1:] {
		reposPath, err := pathutil.NormalizeRepos(arg)
		if err != nil {
			return "", nil, err
		}
		reposPathList = append(reposPathList, reposPath)
	}
	return profileName, reposPathList, nil
}

func (*profileCmd) transactProfile(profileName string, modifyProfile func(*lockjson.Profile)) (*lockjson.LockJSON, error) {
	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		return nil, errors.New("failed to read lock.json: " + err.Error())
	}

	// Return error if profiles[]/name does not match profileName
	profile, err := lockJSON.Profiles.FindByName(profileName)
	if err != nil {
		return nil, err
	}

	// Begin transaction
	err = transaction.Create()
	if err != nil {
		return nil, err
	}
	defer transaction.Remove()

	modifyProfile(profile)

	// Write to lock.json
	err = lockJSON.Write()
	if err != nil {
		return nil, err
	}
	return lockJSON, nil
}
