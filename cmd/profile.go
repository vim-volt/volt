package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/pathutil"
	"github.com/vim-volt/volt/transaction"
)

type profileFlagsType struct {
	helped bool
}

var profileFlags profileFlagsType

type profileCmd struct{}

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
	profileSubCmd["use"] = cmd.doUse

	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Println(`
Usage
  profile [get]
    Get current profile name.

  profile set {name}
    Set profile name to {name}.

  profile show {name}
    Show profile info of {name}.

  profile list
    List all profiles.

  profile new {name}
    Create new profile of {name}. This command does not switch to profile {name}.

  profile destroy {name}
    Delete profile of {name}.
    NOTE: Cannot delete current active profile.

  profile add {name} {repository} [{repository2} ...]
    Add one or more repositories to profile {name}.

  profile rm {name} {repository} [{repository2} ...]
    Remove one or more repositories from profile {name}.

  profile use [-current | {name}] vimrc [true | false]
  profile use [-current | {name}] gvimrc [true | false]
    Set vimrc / gvimrc flag to true or false.

Quick example
  $ volt profile list
  * default
  $ volt profile new foo # will create profile "foo"
  $ volt profile list
  * default
    foo
  $ volt profile set foo # will switch profile to "foo"
  $ volt profile list
    default
  * foo

  $ volt profile use -current vimrc false # Disable installing vimrc on current active profile on "volt rebuild"
  $ volt profile use default gvimrc true  # Enable installing gvimrc on profile default on "volt rebuild"

  $ volt profile set default # will switch profile to "default"
  $ volt profile rm foo tyru/caw.vim # will disable loading tyru/caw.vim plugin on profile "foo"
  $ volt disable tyru/open-browser.vim # if profile name is current active profile, you can also use "volt disable" command instead
  $ volt profile add foo tyru/caw.vim # will enable loading tyru/caw.vim plugin on profile "foo"
  $ volt enable tyru/open-browser.vim
  $ volt profile destroy foo # will delete profile "foo"
`)
		profileFlags.helped = true
	}

	cmdFlagSet["profile"] = fs
}

func Profile(args []string) int {
	cmd := profileCmd{}

	// Parse args
	args, err := cmd.parseArgs(args)
	if err == ErrShowedHelp {
		return 0
	}
	if err != nil {
		logger.Error(err.Error())
		return 10
	}

	if fn, exists := profileSubCmd[args[0]]; exists {
		err = fn(args[1:])
		if err != nil {
			logger.Error(err.Error())
			return 11
		}
	}

	return 0
}

func (cmd *profileCmd) parseArgs(args []string) ([]string, error) {
	fs := cmdFlagSet["profile"]
	fs.Parse(args)
	if profileFlags.helped {
		return nil, ErrShowedHelp
	}

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
		cmdFlagSet["profile"].Usage()
		logger.Error("'volt profile set' receives profile name.")
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
		logger.Info("Unchanged active profile '" + profileName + "'")
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

	logger.Info("Set active profile to '" + profileName + "'")

	// Rebuild start dir
	err = (&rebuildCmd{}).doRebuild(false)
	if err != nil {
		return errors.New("could not rebuild " + pathutil.VimVoltDir() + ": " + err.Error())
	}

	return nil
}

func (cmd *profileCmd) doShow(args []string) error {
	if len(args) == 0 {
		cmdFlagSet["profile"].Usage()
		logger.Error("'volt profile show' receives profile name.")
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
	fmt.Println("use vimrc:", profile.UseVimrc)
	fmt.Println("use gvimrc:", profile.UseGvimrc)
	fmt.Println("repos path:")
	for _, reposPath := range profile.ReposPath {
		hash, err := getRemoteHEAD(reposPath)
		if err != nil {
			hash = "?"
		}
		fmt.Printf("  %s (%s)\n", reposPath, hash)
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
		cmdFlagSet["profile"].Usage()
		logger.Error("'volt profile new' receives profile name.")
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
		Name:      profileName,
		ReposPath: make([]string, 0),
		UseVimrc:  true,
		UseGvimrc: true,
	})

	// Write to lock.json
	err = lockJSON.Write()
	if err != nil {
		return err
	}

	logger.Info("Created new profile '" + profileName + "'")

	return nil
}

func (cmd *profileCmd) doDestroy(args []string) error {
	if len(args) == 0 {
		cmdFlagSet["profile"].Usage()
		logger.Error("'volt profile destroy' receives profile name.")
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

	logger.Info("Deleted profile '" + profileName + "'")

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
				logger.Warn("repository '" + reposPath + "' already exists")
			} else {
				profile.ReposPath = append(profile.ReposPath, reposPath)
				logger.Info("Activate '" + reposPath + "' on profile '" + profileName + "'")
			}
		}
	})
	if err != nil {
		return err
	}

	if lockJSON.ActiveProfile == profileName {
		// Rebuild start dir
		err = (&rebuildCmd{}).doRebuild(false)
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
				logger.Info("Deactivate '" + reposPath + "' from profile '" + profileName + "'")
			} else {
				logger.Warn("repository '" + reposPath + "' does not exist")
			}
		}
	})
	if err != nil {
		return err
	}

	if lockJSON.ActiveProfile == profileName {
		// Rebuild start dir
		err = (&rebuildCmd{}).doRebuild(false)
		if err != nil {
			return errors.New("could not rebuild " + pathutil.VimVoltDir() + ": " + err.Error())
		}
	}

	return nil
}

func (cmd *profileCmd) parseAddArgs(subCmd string, args []string) (string, []string, error) {
	if len(args) == 0 {
		cmdFlagSet["profile"].Usage()
		logger.Errorf("'volt profile %s' receives profile name and one or more repositories.", subCmd)
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

func (cmd *profileCmd) doUse(args []string) error {
	// Validate arguments
	if len(args) != 3 {
		cmdFlagSet["profile"].Usage()
		logger.Error("'volt profile use' receives profile name, rc name, value.")
		return nil
	}
	if args[1] != "vimrc" && args[1] != "gvimrc" {
		cmdFlagSet["profile"].Usage()
		logger.Error("volt profile use: Please specify \"vimrc\" or \"gvimrc\" to the 2nd argument")
		return nil
	}
	if args[2] != "true" && args[2] != "false" {
		cmdFlagSet["profile"].Usage()
		logger.Error("volt profile use: Please specify \"true\" or \"false\" to the 3rd argument")
		return nil
	}

	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		return errors.New("failed to read lock.json: " + err.Error())
	}

	// Convert arguments
	var profileName string
	var rcName string
	var value bool
	if args[0] == "-current" {
		profileName = lockJSON.ActiveProfile
	} else {
		profileName = args[0]
	}
	rcName = args[1]
	if args[2] == "true" {
		value = true
	} else {
		value = false
	}

	// Look up specified profile
	profile, err := lockJSON.Profiles.FindByName(profileName)
	if err != nil {
		return err
	}

	// Begin transaction
	err = transaction.Create()
	if err != nil {
		return err
	}
	defer transaction.Remove()

	// Set use_vimrc / use_gvimrc flag
	changed := false
	if rcName == "vimrc" {
		if profile.UseVimrc != value {
			logger.Infof("Set vimrc flag of profile '%s' to '%s'", profileName, strconv.FormatBool(value))
			profile.UseVimrc = value
			changed = true
		} else {
			logger.Warnf("vimrc flag of profile '%s' is already '%s'", profileName, strconv.FormatBool(value))
		}
	} else {
		if profile.UseGvimrc != value {
			logger.Infof("Set gvimrc flag of profile '%s' to '%s'", profileName, strconv.FormatBool(value))
			profile.UseGvimrc = value
			changed = true
		} else {
			logger.Warnf("gvimrc flag of profile '%s' is already '%s'", profileName, strconv.FormatBool(value))
		}
	}

	if changed {
		// Write to lock.json
		err = lockJSON.Write()
		if err != nil {
			return err
		}

		// Rebuild start dir
		err = (&rebuildCmd{}).doRebuild(false)
		if err != nil {
			return errors.New("could not rebuild " + pathutil.VimVoltDir() + ": " + err.Error())
		}
	}

	return nil
}
