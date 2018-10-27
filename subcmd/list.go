package subcmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"text/template"

	"github.com/vim-volt/volt/lockjson"
)

func init() {
	cmdMap["list"] = &listCmd{}
}

type listCmd struct {
	helped bool
	format string
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

func (cmd *listCmd) Run(args []string) *Error {
	fs := cmd.FlagSet()
	fs.Parse(args)
	if cmd.helped {
		return nil
	}
	if err := cmd.list(cmd.format); err != nil {
		return &Error{Code: 10, Msg: "Failed to render template: " + err.Error()}
	}
	return nil
}

func (cmd *listCmd) list(format string) error {
	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		return errors.Wrap(err, "failed to read lock.json")
	}
	// Parse template string
	t, err := template.New("volt").Funcs(cmd.funcMap(lockJSON)).Parse(format)
	if err != nil {
		return err
	}
	// Output templated information
	return t.Execute(os.Stdout, lockJSON)
}

func (*listCmd) funcMap(lockJSON *lockjson.LockJSON) template.FuncMap {
	profileOf := func(name string) *lockjson.Profile {
		profile, err := lockJSON.Profiles.FindByName(name)
		if err != nil {
			return &lockjson.Profile{}
		}
		return profile
	}

	return template.FuncMap{
		"json": func(value interface{}, args ...string) string {
			var b []byte
			switch len(args) {
			case 0:
				b, _ = json.MarshalIndent(value, "", "")
			case 1:
				b, _ = json.MarshalIndent(value, args[0], "")
			default:
				b, _ = json.MarshalIndent(value, args[0], args[1])
			}
			return string(b)
		},
		"currentProfile": func() *lockjson.Profile {
			return profileOf(lockJSON.CurrentProfileName)
		},
		"profile": profileOf,
		"version": func() string {
			return voltVersion
		},
		"versionMajor": func() int {
			return voltVersionInfo()[0]
		},
		"versionMinor": func() int {
			return voltVersionInfo()[1]
		},
		"versionPatch": func() int {
			return voltVersionInfo()[2]
		},
	}
}
