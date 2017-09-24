package lockjson

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/vim-volt/go-volt/pathutil"
)

type LockJSON struct {
	Version       int64     `json:"version"`
	TrxID         int64     `json:"trx_id"`
	ActiveProfile string    `json:"active_profile"`
	LoadInit      bool      `json:"load_init"`
	Repos         []Repos   `json:"repos"`
	Profiles      []Profile `json:"profiles"`
}

type Repos struct {
	TrxID   int64  `json:"trx_id"`
	Path    string `json:"path"`
	Version string `json:"version"`
	Active  bool   `json:"active"`
}

type Profile struct {
	Name      string   `json:"name"`
	ReposPath []string `json:"repos_path"`
	LoadInit  bool     `json:"load_init"`
}

func InitialLockJSON() *LockJSON {
	return &LockJSON{
		Version:       1,
		TrxID:         1,
		ActiveProfile: "default",
		LoadInit:      true,
		Repos:         make([]Repos, 0),
		Profiles:      make([]Profile, 0),
	}
}

func Read() (*LockJSON, error) {
	// Return initial lock.json struct if lockfile does not exist
	lockfile := pathutil.LockJSON()
	if _, err := os.Stat(lockfile); os.IsNotExist(err) {
		return InitialLockJSON(), nil
	}

	// Read lock.json
	bytes, err := ioutil.ReadFile(lockfile)
	if err != nil {
		return nil, err
	}
	var lockJSON LockJSON
	err = json.Unmarshal(bytes, &lockJSON)
	if err != nil {
		return nil, err
	}

	// Validate lock.json
	err = validate(&lockJSON)
	if err != nil {
		return nil, err
	}

	return &lockJSON, nil
}

func validate(lockJSON *LockJSON) error {
	// Validate if missing required keys exist
	err := validateMissing(lockJSON)
	if err != nil {
		return err
	}

	// Validate if duplicate repos[]/path exist
	dup := make(map[string]bool, len(lockJSON.Repos))
	for _, repos := range lockJSON.Repos {
		if _, exists := dup[repos.Path]; exists {
			return errors.New("duplicate repos: " + repos.Path)
		}
		dup[repos.Path] = true
	}

	// Validate if duplicate profiles[]/name exist
	dup = make(map[string]bool, len(lockJSON.Profiles))
	for _, profile := range lockJSON.Profiles {
		if _, exists := dup[profile.Name]; exists {
			return errors.New("duplicate profile: " + profile.Name)
		}
		dup[profile.Name] = true
	}

	// Validate if duplicate profiles[]/repos_path[] exist
	dup = make(map[string]bool, len(lockJSON.Profiles)*10)
	for _, profile := range lockJSON.Profiles {
		for _, reposPath := range profile.ReposPath {
			if _, exists := dup[reposPath]; exists {
				return errors.New("duplicate 'repos_path' (" + reposPath + ") in profile '" + profile.Name + "'")
			}
			dup[reposPath] = true
		}
	}

	// Validate if active_profile exists in profiles[]/name
	if lockJSON.ActiveProfile != "default" {
		found := false
		for _, profile := range lockJSON.Profiles {
			if profile.Name == lockJSON.ActiveProfile {
				found = true
				break
			}
		}
		if !found {
			return errors.New("'active_profile' (" + lockJSON.ActiveProfile + ") doesn't exist in profiles")
		}
	}

	// Validate if profiles[]/repos_path[] exists in repos[]/path
	for i, profile := range lockJSON.Profiles {
		for j, reposPath := range profile.ReposPath {
			found := false
			for _, repos := range lockJSON.Repos {
				if reposPath == repos.Path {
					found = true
					break
				}
			}
			if !found {
				return errors.New(
					"'profiles[" + strconv.Itoa(i) + "].repos_path[" + strconv.Itoa(j) +
						"]' (" + reposPath + ") doesn't exist in repos")
			}
		}
	}

	// Validate if repos[]/path exists on filesystem
	// and is a directory
	for i, repos := range lockJSON.Repos {
		fullpath := pathutil.FullReposPathOf(repos.Path)
		if file, err := os.Stat(fullpath); os.IsNotExist(err) {
			return errors.New("'repos[" + strconv.Itoa(i) + "].path' (" + fullpath + ") doesn't exist on filesystem")
		} else if !file.IsDir() {
			return errors.New("'repos[" + strconv.Itoa(i) + "].path' (" + fullpath + ") is not a directory")
		}
	}

	// Validate if trx_id is equal or greater than repos[]/trx_id
	index := -1
	var max int64
	for i, repos := range lockJSON.Repos {
		if max < repos.TrxID {
			index = i
			max = repos.TrxID
		}
	}
	if max > lockJSON.TrxID {
		return errors.New("'repos[" + strconv.Itoa(index) + "].trx_id' (" + strconv.FormatInt(max, 10) + ") is greater than 'trx_id' (" + strconv.FormatInt(lockJSON.TrxID, 10) + ")")
	}

	return nil
}

func validateMissing(lockJSON *LockJSON) error {
	if lockJSON.Version == 0 {
		return errors.New("missing 'version'")
	}
	if lockJSON.TrxID == 0 {
		return errors.New("missing 'trx_id'")
	}
	if lockJSON.Repos == nil {
		return errors.New("missing 'repos'")
	}
	for i, repos := range lockJSON.Repos {
		if repos.TrxID == 0 {
			return errors.New("missing 'repos[" + strconv.Itoa(i) + "].trx_id'")
		}
		if repos.Path == "" {
			return errors.New("missing 'repos[" + strconv.Itoa(i) + "].path'")
		}
		if repos.Version == "" {
			return errors.New("missing 'repos[" + strconv.Itoa(i) + "].version'")
		}
	}
	if lockJSON.Profiles == nil {
		return errors.New("missing 'profiles'")
	}
	for i, profile := range lockJSON.Profiles {
		if profile.Name == "" {
			return errors.New("missing 'profile[" + strconv.Itoa(i) + "].name'")
		}
		if profile.ReposPath == nil {
			return errors.New("missing 'profile[" + strconv.Itoa(i) + "].repos_path'")
		}
		for j, reposPath := range profile.ReposPath {
			if reposPath == "" {
				return errors.New("missing 'profile[" + strconv.Itoa(i) + "].repos_path[" + strconv.Itoa(j) + "]'")
			}
		}
	}
	return nil
}

func Write(lockJSON *LockJSON) error {
	// Mkdir all if lock.json's directory does not exist
	lockfile := pathutil.LockJSON()
	if _, err := os.Stat(filepath.Dir(lockfile)); os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(lockfile), 0755)
		if err != nil {
			return err
		}
	}

	// Write to lock.json
	bytes, err := json.MarshalIndent(lockJSON, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(pathutil.LockJSON(), bytes, 0644)
}
