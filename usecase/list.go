package usecase

import (
	"encoding/json"
	"html/template"
	"io"

	"github.com/vim-volt/volt/config"
	"github.com/vim-volt/volt/lockjson"
)

// List renders text/template format format to w with paramter lockJSON, cfg.
func List(w io.Writer, format string, lockJSON *lockjson.LockJSON, cfg *config.Config) error {
	// Parse template string
	t, err := template.New("volt").Funcs(funcMap(lockJSON)).Parse(format)
	if err != nil {
		return err
	}
	// Output templated information
	return t.Execute(w, lockJSON)
}

func funcMap(lockJSON *lockjson.LockJSON) template.FuncMap {
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
			return VersionString()
		},
		"versionMajor": func() int {
			return Version()[0]
		},
		"versionMinor": func() int {
			return Version()[1]
		},
		"versionPatch": func() int {
			return Version()[2]
		},
	}
}
