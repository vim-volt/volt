package gateway

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/vim-volt/volt/config"
	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/usecase"
)

func init() {
	cmdMap["list"] = &listCmd{
		List: usecase.List,
	}
}

type listCmd struct {
	helped bool
	format string

	List func(w io.Writer, format string, lockJSON *lockjson.LockJSON, cfg *config.Config) error
}

func (cmd *listCmd) ProhibitRootExecution(args []string) bool { return false }

func (cmd *listCmd) FlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Print(`
Usage
  volt list [-help] [-f {text/template string}]

Quick example
  $ volt list # will list installed plugins

  Show all installed repositories:

  $ volt list -f '{{ range .Repos }}{{ println .Path }}{{ end }}'

  Show repositories used by current profile:

  $ volt list -f '{{ range .Profiles }}{{ if eq $.CurrentProfileName .Name }}{{ range .ReposPath }}{{ println . }}{{ end }}{{ end }}{{ end }}'

  Or (see "Additional property"):

  $ volt list -f '{{ range currentProfile.ReposPath }}{{ println . }}{{ end }}'

Template functions

  json value [prefix [indent]] (string)
    Returns JSON representation of value.
    The argument is same as json.MarshalIndent().

  currentProfile (Profile (see "Structures"))
    Returns current profile

  currentProfile (Profile (see "Structures"))
    Returns given name's profile

  version (string)
    Returns volt version string. format is "v{major}.{minor}.{patch}" (e.g. "v0.3.0")

  versionMajor (number)
    Returns volt major version

  versionMinor (number)
    Returns volt minor version

  versionPatch (number)
    Returns volt patch version

Structures
  This describes the structure of lock.json .
  {
    // lock.json structure compatibility version
    "version": <int64>,

    // Current profile name (e.g. "default")
    "current_profile_name": <string>,

    // All Installed repositories
    // ("volt list" shows current profile's repositories, which is not the same as this)
    "repos": [
      {
        // "git" (git repository) or "static" (static repository)
        "type": <string>,

        // Repository path like "github.com/vim-volt/vim-volt"
        "path": <string>,

        // Git commit hash. if "type" is "static" this property does not exist
        "version": <string>,
      },
    ],

    // Profiles
    "profiles": [
      // Profile name (.e.g. "default")
      "name": <string>,

      // Repositories ("volt list" shows these repositories)
      "repos_path": [ <string> ],
    ]
  }

Description
  Vim plugin information extractor.
  If -f flag is not given, this command shows vim plugins of **current profile** (not all installed plugins) by default.
  If -f flag is given, it renders by given template which can access the information of lock.json .` + "\n\n")
		//fmt.Println("Options")
		//fs.PrintDefaults()
		fmt.Println()
		cmd.helped = true
	}
	fs.StringVar(&cmd.format, "f", cmd.defaultTemplate(), "text/template format string")
	return fs
}

func (*listCmd) defaultTemplate() string {
	return `name: {{ .CurrentProfileName }}
repos path:
{{- range currentProfile.ReposPath }}
  {{ . }}
{{- end }}
`
}

func (cmd *listCmd) Run(cmdctx *CmdContext) *Error {
	fs := cmd.FlagSet()
	fs.Parse(cmdctx.Args)
	if cmd.helped {
		return nil
	}
	if err := cmd.List(os.Stdout, cmd.format, cmdctx.LockJSON, cmdctx.Config); err != nil {
		return &Error{Code: 10, Msg: "Failed to render template: " + err.Error()}
	}
	return nil
}
