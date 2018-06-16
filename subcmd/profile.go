package subcmd

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/hashicorp/go-multierror"
	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/pathutil"
	"github.com/vim-volt/volt/subcmd/builder"
	"github.com/vim-volt/volt/transaction"
)

type profileCmd struct {
	helped bool
}

var profileSubCmd = make(map[string]func([]string) error)

func init() {
	cmdMap["profile"] = &profileCmd{}
}

func (cmd *profileCmd) ProhibitRootExecution(args []string) bool {
	if len(args) == 0 {
		return true
	}
	subCmd := args[0]
	switch subCmd {
	case "show":
		return false
	case "list":
		return false
	default:
		return true
	}
}

func (cmd *profileCmd) FlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Print(`
Usage
  profile [-help] {command}

Command
  profile set [-n] {name}
    Set profile name to {name}.

  profile show [-current | {name}]
    Show profile info of {name}.

  profile list
    List all profiles.

  profile new {name}
    Create new profile of {name}. This command does not switch to profile {name}.

  profile destroy {name}
    Delete profile of {name}.
    NOTE: Cannot delete current profile.

  profile rename {old} {new}
    Rename profile {old} to {new}.

  profile add [-current | {name}] {repository} [{repository2} ...]
    Add one or more repositories to profile {name}.

  profile rm [-current | {name}] {repository} [{repository2} ...]
    Remove one or more repositories from profile {name}.

Quick example
  $ volt profile list   # default profile is "default"
  * default
  $ volt profile new foo   # will create profile "foo"
  $ volt profile list
  * default
    foo
  $ volt profile set foo   # will switch profile to "foo"
  $ volt profile list
    default
  * foo

  $ volt profile set default   # on profile "default"

  $ volt enable tyru/caw.vim    # enable loading tyru/caw.vim on current profile
  $ volt profile add foo tyru/caw.vim    # enable loading tyru/caw.vim on "foo" profile

  $ volt disable tyru/caw.vim   # disable loading tyru/caw.vim on current profile
  $ volt profile rm foo tyru/caw.vim    # disable loading tyru/caw.vim on "foo" profile

  $ volt profile destroy foo   # will delete profile "foo"` + "\n\n")
		cmd.helped = true
	}
	return fs
}

func (cmd *profileCmd) Run(runctx *RunContext) *Error {
	// Parse args
	args, err := cmd.parseArgs(runctx.Args)
	if err == ErrShowedHelp {
		return nil
	}
	if err != nil {
		return &Error{Code: 10, Msg: err.Error()}
	}

	subCmd := args[0]
	runctx.Args = args[1:]
	switch subCmd {
	case "set":
		err = cmd.doSet(runctx)
	case "show":
		err = cmd.doShow(runctx)
	case "list":
		err = cmd.doList(runctx)
	case "new":
		err = cmd.doNew(runctx)
	case "destroy":
		err = cmd.doDestroy(runctx)
	case "rename":
		err = cmd.doRename(runctx)
	case "add":
		err = cmd.doAdd(runctx)
	case "rm":
		err = cmd.doRm(runctx)
	default:
		return &Error{Code: 11, Msg: "Unknown subcommand: " + subCmd}
	}

	if err != nil {
		return &Error{Code: 20, Msg: err.Error()}
	}

	return nil
}

func (cmd *profileCmd) parseArgs(args []string) ([]string, error) {
	fs := cmd.FlagSet()
	fs.Parse(args)
	if cmd.helped {
		return nil, ErrShowedHelp
	}
	if len(fs.Args()) == 0 {
		fs.Usage()
		logger.Error("must specify subcommand")
		return nil, ErrShowedHelp
	}
	return fs.Args(), nil
}

func (cmd *profileCmd) doSet(runctx *RunContext) (result error) {
	args := runctx.Args
	lockJSON := runctx.LockJSON

	// Parse args
	createProfile := false
	if len(args) > 0 && args[0] == "-n" {
		createProfile = true
		args = args[1:]
	}
	if len(args) == 0 {
		cmd.FlagSet().Usage()
		logger.Error("'volt profile set' receives profile name.")
		return nil
	}
	profileName := args[0]

	// Exit if current profile is same as profileName
	if lockJSON.CurrentProfileName == profileName {
		return fmt.Errorf("'%s' is current profile", profileName)
	}

	// Create given profile unless the profile exists
	if _, err := lockJSON.Profiles.FindByName(profileName); err != nil {
		if !createProfile {
			return err
		}
		runctx.Args = []string{profileName}
		if err := cmd.doNew(runctx); err != nil {
			return err
		}
		if _, err = lockJSON.Profiles.FindByName(profileName); err != nil {
			return err
		}
	}

	// Begin transaction
	trx, err := transaction.Start()
	if err != nil {
		return err
	}
	defer func() {
		if err := trx.Done(); err != nil {
			result = err
		}
	}()

	// Set profile name
	lockJSON.CurrentProfileName = profileName

	// Write to lock.json
	err = lockJSON.Write()
	if err != nil {
		return err
	}

	logger.Info("Changed current profile: " + profileName)

	// Build ~/.vim/pack/volt dir
	err = builder.Build(false, lockJSON, runctx.Config)
	if err != nil {
		return errors.New("could not build " + pathutil.VimVoltDir() + ": " + err.Error())
	}

	return nil
}

func (cmd *profileCmd) doShow(runctx *RunContext) error {
	args := runctx.Args
	lockJSON := runctx.LockJSON

	if len(args) == 0 {
		cmd.FlagSet().Usage()
		logger.Error("'volt profile show' receives profile name.")
		return nil
	}

	var profileName string
	if args[0] == "-current" {
		profileName = lockJSON.CurrentProfileName
	} else {
		profileName = args[0]
		if lockJSON.Profiles.FindIndexByName(profileName) == -1 {
			return fmt.Errorf("profile '%s' does not exist", profileName)
		}
	}

	return (&listCmd{}).list(fmt.Sprintf(`name: %s
repos path:
{{- with profile %q -}}
{{- range .ReposPath }}
  {{ . }}
{{- end -}}
{{- end }}
`, profileName, profileName), lockJSON)
}

func (cmd *profileCmd) doList(runctx *RunContext) error {
	return (&listCmd{}).list(`
{{- range .Profiles -}}
{{- if eq .Name $.CurrentProfileName -}}*{{- else }} {{ end }} {{ .Name }}
{{ end -}}
`, runctx.LockJSON)
}

func (cmd *profileCmd) doNew(runctx *RunContext) (result error) {
	args := runctx.Args
	lockJSON := runctx.LockJSON

	if len(args) == 0 {
		cmd.FlagSet().Usage()
		logger.Error("'volt profile new' receives profile name.")
		return nil
	}
	profileName := args[0]

	// Return error if profiles[]/name matches profileName
	_, err := lockJSON.Profiles.FindByName(profileName)
	if err == nil {
		return errors.New("profile '" + profileName + "' already exists")
	}

	// Begin transaction
	trx, err := transaction.Start()
	if err != nil {
		return err
	}
	defer func() {
		if err := trx.Done(); err != nil {
			result = err
		}
	}()

	// Add profile
	lockJSON.Profiles = append(lockJSON.Profiles, lockjson.Profile{
		Name:      profileName,
		ReposPath: make([]pathutil.ReposPath, 0),
	})

	// Write to lock.json
	err = lockJSON.Write()
	if err != nil {
		return err
	}

	logger.Info("Created new profile '" + profileName + "'")

	return nil
}

func (cmd *profileCmd) doDestroy(runctx *RunContext) (result error) {
	args := runctx.Args
	lockJSON := runctx.LockJSON

	if len(args) == 0 {
		cmd.FlagSet().Usage()
		logger.Error("'volt profile destroy' receives profile name.")
		return nil
	}

	// Begin transaction
	trx, err := transaction.Start()
	if err != nil {
		return err
	}
	defer func() {
		if err := trx.Done(); err != nil {
			result = err
		}
	}()

	var merr *multierror.Error
	for i := range args {
		profileName := args[i]

		// Skip if current profile matches profileName
		if lockJSON.CurrentProfileName == profileName {
			merr = multierror.Append(merr, errors.New("cannot destroy current profile: "+profileName))
			continue
		}
		// Skip if profiles[]/name does not match profileName
		index := lockJSON.Profiles.FindIndexByName(profileName)
		if index < 0 {
			merr = multierror.Append(merr, errors.New("profile '"+profileName+"' does not exist"))
			continue
		}

		// Remove the specified profile
		lockJSON.Profiles = append(lockJSON.Profiles[:index], lockJSON.Profiles[index+1:]...)

		// Remove $VOLTPATH/rc/{profile} dir
		rcDir := pathutil.RCDir(profileName)
		os.RemoveAll(rcDir)
		if pathutil.Exists(rcDir) {
			return errors.New("failed to remove " + rcDir)
		}

		logger.Info("Deleted profile '" + profileName + "'")
	}

	// Write to lock.json
	err = lockJSON.Write()
	if err != nil {
		return err
	}

	return merr.ErrorOrNil()
}

func (cmd *profileCmd) doRename(runctx *RunContext) (result error) {
	args := runctx.Args
	lockJSON := runctx.LockJSON

	if len(args) != 2 {
		cmd.FlagSet().Usage()
		logger.Error("'volt profile rename' receives profile name.")
		return nil
	}
	oldName := args[0]
	newName := args[1]

	// Return error if profiles[]/name does not match oldName
	index := lockJSON.Profiles.FindIndexByName(oldName)
	if index < 0 {
		return errors.New("profile '" + oldName + "' does not exist")
	}

	// Return error if profiles[]/name does not match newName
	if lockJSON.Profiles.FindIndexByName(newName) >= 0 {
		return errors.New("profile '" + newName + "' already exists")
	}

	// Begin transaction
	trx, err := transaction.Start()
	if err != nil {
		return err
	}
	defer func() {
		if err := trx.Done(); err != nil {
			result = err
		}
	}()

	// Rename profile names
	lockJSON.Profiles[index].Name = newName
	if lockJSON.CurrentProfileName == oldName {
		lockJSON.CurrentProfileName = newName
	}

	// Rename $VOLTPATH/rc/{profile} dir
	oldRCDir := pathutil.RCDir(oldName)
	if pathutil.Exists(oldRCDir) {
		newRCDir := pathutil.RCDir(newName)
		if err = os.Rename(oldRCDir, newRCDir); err != nil {
			return fmt.Errorf("could not rename %s to %s", oldRCDir, newRCDir)
		}
	}

	// Write to lock.json
	err = lockJSON.Write()
	if err != nil {
		return err
	}

	logger.Infof("Renamed profile '%s' to '%s'", oldName, newName)

	return nil
}

func (cmd *profileCmd) doAdd(runctx *RunContext) error {
	args := runctx.Args
	lockJSON := runctx.LockJSON

	// Parse args
	profileName, reposPathList, err := cmd.parseAddArgs(lockJSON, "add", args)
	if err != nil {
		return errors.New("failed to parse args: " + err.Error())
	}

	if profileName == "-current" {
		profileName = lockJSON.CurrentProfileName
	}

	// Read modified profile and write to lock.json
	lockJSON, err = cmd.transactProfile(lockJSON, profileName, func(profile *lockjson.Profile) {
		// Add repositories to profile if the repository does not exist
		for _, reposPath := range reposPathList {
			if profile.ReposPath.Contains(reposPath) {
				logger.Warn("repository '" + reposPath.String() + "' is already enabled")
			} else {
				profile.ReposPath = append(profile.ReposPath, reposPath)
				logger.Info("Enabled '" + reposPath.String() + "' on profile '" + profileName + "'")
			}
		}
	})
	if err != nil {
		return err
	}

	// Build ~/.vim/pack/volt dir
	err = builder.Build(false, lockJSON, runctx.Config)
	if err != nil {
		return errors.New("could not build " + pathutil.VimVoltDir() + ": " + err.Error())
	}

	return nil
}

func (cmd *profileCmd) doRm(runctx *RunContext) error {
	args := runctx.Args
	lockJSON := runctx.LockJSON

	// Parse args
	profileName, reposPathList, err := cmd.parseAddArgs(lockJSON, "rm", args)
	if err != nil {
		return errors.New("failed to parse args: " + err.Error())
	}

	if profileName == "-current" {
		profileName = lockJSON.CurrentProfileName
	}

	// Read modified profile and write to lock.json
	lockJSON, err = cmd.transactProfile(lockJSON, profileName, func(profile *lockjson.Profile) {
		// Remove repositories from profile if the repository does not exist
		for _, reposPath := range reposPathList {
			index := profile.ReposPath.IndexOf(reposPath)
			if index >= 0 {
				// Remove profile.ReposPath[index]
				profile.ReposPath = append(profile.ReposPath[:index], profile.ReposPath[index+1:]...)
				logger.Info("Disabled '" + reposPath.String() + "' from profile '" + profileName + "'")
			} else {
				logger.Warn("repository '" + reposPath.String() + "' is already disabled")
			}
		}
	})
	if err != nil {
		return err
	}

	// Build ~/.vim/pack/volt dir
	err = builder.Build(false, lockJSON, runctx.Config)
	if err != nil {
		return errors.New("could not build " + pathutil.VimVoltDir() + ": " + err.Error())
	}

	return nil
}

func (cmd *profileCmd) parseAddArgs(lockJSON *lockjson.LockJSON, subCmd string, args []string) (string, []pathutil.ReposPath, error) {
	if len(args) == 0 {
		cmd.FlagSet().Usage()
		logger.Errorf("'volt profile %s' receives profile name and one or more repositories.", subCmd)
		return "", nil, nil
	}

	profileName := args[0]
	reposPathList := make([]pathutil.ReposPath, 0, len(args)-1)
	for _, arg := range args[1:] {
		reposPath, err := pathutil.NormalizeRepos(arg)
		if err != nil {
			return "", nil, err
		}
		reposPathList = append(reposPathList, reposPath)
	}

	// Validate if all repositories exist in repos[]
	for i := range reposPathList {
		_, err := lockJSON.Repos.FindByPath(reposPathList[i])
		if err != nil {
			return "", nil, err
		}
	}

	return profileName, reposPathList, nil
}

// Run modifyProfile and write modified structure to lock.json
func (*profileCmd) transactProfile(lockJSON *lockjson.LockJSON, profileName string, modifyProfile func(*lockjson.Profile)) (resultLockJSON *lockjson.LockJSON, result error) {
	// Return error if profiles[]/name does not match profileName
	profile, err := lockJSON.Profiles.FindByName(profileName)
	if err != nil {
		return nil, err
	}

	// Begin transaction
	trx, err := transaction.Start()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := trx.Done(); err != nil {
			result = err
		}
	}()

	modifyProfile(profile)

	// Write to lock.json
	err = lockJSON.Write()
	if err != nil {
		return nil, err
	}
	return lockJSON, nil
}
