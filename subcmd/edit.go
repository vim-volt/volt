package subcmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/vim-volt/volt/config"
	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/pathutil"
	"github.com/vim-volt/volt/subcmd/builder"
)

func init() {
	cmdMap["edit"] = &editCmd{}
}

type editCmd struct {
	helped bool
	editor string
}

func (cmd *editCmd) ProhibitRootExecution(args []string) bool { return true }

func (cmd *editCmd) FlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Print(`
Usage
  volt edit [-help] [-e|--editor {editor}] {repository} [{repository2} ...]

Quick example
  $ volt edit tyru/caw.vim # will open the plugconf file for tyru/caw.vim for editing

Description
  Open the plugconf file(s) of one or more {repository} for editing.

  If the -e option was given, use the given editor for editing those files (unless it cannot be found)

  It also calls "volt build" afterwards if modifications were made to the plugconf file(s).` + "\n\n")
		//fmt.Println("Options")
		//fs.PrintDefaults()
		fmt.Println()
		cmd.helped = true
	}
	fs.StringVar(&cmd.editor, "editor", "", "Use the given editor for editing the plugconf files")
	fs.StringVar(&cmd.editor, "e", "", "Use the given editor for editing the plugconf files")
	return fs
}

func (cmd *editCmd) Run(args []string) *Error {
	reposPathList, err := cmd.parseArgs(args)
	if err == ErrShowedHelp {
		return nil
	}
	if err != nil {
		return &Error{Code: 10, Msg: "Failed to parse args: " + err.Error()}
	}

	hasChanges, err := cmd.doEdit(reposPathList)
	if err != nil {
		return &Error{Code: 15, Msg: "Failed to edit plugconf file: " + err.Error()}
	}

	// Build opt dir
	if hasChanges {
		err = builder.Build(false)
		if err != nil {
			return &Error{Code: 12, Msg: "Could not build " + pathutil.VimVoltDir() + ": " + err.Error()}
		}
	}

	return nil
}

func (cmd *editCmd) doEdit(reposPathList []pathutil.ReposPath) (bool, error) {
	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		return false, err
	}

	// Read config.toml
	cfg, err := config.Read()
	if err != nil {
		return false, errors.New("could not read config.toml: " + err.Error())
	}

	editor, err := cmd.identifyEditor(cfg)
	if err != nil || editor == "" {
		return false, &Error{Code: 30, Msg: "No usable editor found"}
	}

	changeWasMade := false
	for _, reposPath := range reposPathList {

		// Edit plugconf file
		plugconfPath := reposPath.Plugconf()

		// Install a new template if none exists
		if !pathutil.Exists(plugconfPath) {
			getCmd := new(getCmd)
			logger.Debugf("Installing new plugconf for '%s'.", reposPath)
			getCmd.downloadPlugconf(reposPath)
		}

		// Remember modification time before opening the editor
		info, err := os.Stat(plugconfPath)
		if err != nil {
			return false, err
		}
		mTimeBefore := info.ModTime()

		// Call the editor with the plugconf file
		editorCmd := exec.Command(editor, plugconfPath)
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		if err = editorCmd.Run(); err != nil {
			logger.Error("Error calling editor for '%s': %s", reposPath, err.Error)
			continue
		}

		// Get modification time after closing the editor
		info, err = os.Stat(plugconfPath)
		if err != nil {
			return false, err
		}
		mTimeAfter := info.ModTime()

		// A change was made if the modification time was updated
		changeWasMade = changeWasMade || mTimeAfter.After(mTimeBefore)

		// Remove repository from lock.json
		err = lockJSON.Repos.RemoveAllReposPath(reposPath)
		err2 := lockJSON.Profiles.RemoveAllReposPath(reposPath)
		if err == nil || err2 == nil {
			// ignore?
		}
	}

	// Write to lock.json
	if err = lockJSON.Write(); err != nil {
		return changeWasMade, err
	}
	return changeWasMade, nil
}

func (cmd *editCmd) parseArgs(args []string) (pathutil.ReposPathList, error) {
	fs := cmd.FlagSet()
	fs.Parse(args)
	if cmd.helped {
		return nil, ErrShowedHelp
	}

	if len(fs.Args()) == 0 {
		fs.Usage()
		return nil, errors.New("repository was not given")
	}

	// Normalize repos path
	reposPathList := make(pathutil.ReposPathList, 0, len(fs.Args()))
	for _, arg := range fs.Args() {
		reposPath, err := pathutil.NormalizeRepos(arg)
		if err != nil {
			return nil, err
		}
		reposPathList = append(reposPathList, reposPath)
	}

	return reposPathList, nil
}

func (cmd *editCmd) identifyEditor(cfg *config.Config) (string, error) {
	editors := make([]string, 0, 6)

	// if an editor is specified as commandline argument, consider it
	// as alternative
	if cmd.editor != "" {
		editors = append(editors, cmd.editor)
	}

	// if an editor is configured in the config.toml, consider it as
	// alternative
	if cfg.Edit.Editor != "" {
		editors = append(editors, cfg.Edit.Editor)
	}

	vimExecutable, err := pathutil.VimExecutable()
	if err != nil {
		logger.Debug("No vim executable found in $PATH")
	} else {
		editors = append(editors, vimExecutable)
	}

	// specifiy a fixed list of other alternatives
	editors = append(editors, "$VISUAL", "sensible-editor", "$EDITOR")

	for _, editor := range editors {
		// resolve content of environment variables
		var editorName string
		if editor[0] == '$' {
			editorName = os.Getenv(editor[1:])
		} else {
			editorName = editor
		}

		path, err := exec.LookPath(editorName)
		if err != nil {
			logger.Debug(editorName + " not found in $PATH")
		} else if path != "" {
			logger.Debug("Using " + path + " as editor")
			return editorName, nil
		}
	}

	return "", errors.New("No usable editor found")
}
