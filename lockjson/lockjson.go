package lockjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/pathutil"
)

type ReposList []Repos
type ProfileList []Profile

type LockJSON struct {
	Version            int64       `json:"version"`
	TrxID              int64       `json:"trx_id"`
	CurrentProfileName string      `json:"current_profile_name"`
	Repos              ReposList   `json:"repos"`
	Profiles           ProfileList `json:"profiles"`
}

type ReposType string

const (
	ReposGitType    ReposType = "git"
	ReposStaticType ReposType = "static"
)

type Repos struct {
	Type    ReposType `json:"type"`
	TrxID   int64     `json:"trx_id"`
	Path    string    `json:"path"`
	Version string    `json:"version"`
}

type profReposPath []string

type Profile struct {
	Name      string        `json:"name"`
	ReposPath profReposPath `json:"repos_path"`
	UseVimrc  bool          `json:"use_vimrc"`
	UseGvimrc bool          `json:"use_gvimrc"`
}

const lockJSONVersion = 2

func InitialLockJSON() *LockJSON {
	return &LockJSON{
		Version:            lockJSONVersion,
		TrxID:              1,
		CurrentProfileName: "default",
		Repos:              make([]Repos, 0),
		Profiles: []Profile{
			Profile{
				Name:      "default",
				ReposPath: make([]string, 0),
				UseVimrc:  true,
				UseGvimrc: true,
			},
		},
	}
}

func Read() (*LockJSON, error) {
	// Return initial lock.json struct if lockfile does not exist
	lockfile := pathutil.LockJSON()
	if !pathutil.Exists(lockfile) {
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

	if lockJSON.Version < lockJSONVersion {
		logger.Warnf("Performing auto-migration of lock.json: v%d -> v%d", lockJSON.Version, lockJSONVersion)
		logger.Warn("Please run 'volt migrate' to migrate explicitly if it's not updated by after operations")
		err = migrate(bytes, &lockJSON)
		if err != nil {
			return nil, err
		}
	}

	// Validate lock.json
	err = validate(&lockJSON)
	if err != nil {
		return nil, errors.New("validation failed: lock.json: " + err.Error())
	}

	return &lockJSON, nil
}

func validate(lockJSON *LockJSON) error {
	if lockJSON.Version < 1 {
		return fmt.Errorf("lock.json version is '%d' (must be 1 or greater)", lockJSON.Version)
	}
	// Validate if volt can manipulate lock.json of this version
	if lockJSON.Version > lockJSONVersion {
		return fmt.Errorf("this lock.json version is '%d' which volt cannot recognize. please upgrade volt to process this file", lockJSON.Version)
	}

	// Validate if missing required keys exist
	err := validateMissing(lockJSON)
	if err != nil {
		return err
	}

	// Validate if duplicate repos[]/path exist
	dup := make(map[string]bool, len(lockJSON.Repos))
	for i := range lockJSON.Repos {
		repos := &lockJSON.Repos[i]
		if _, exists := dup[repos.Path]; exists {
			return errors.New("duplicate repos '" + repos.Path + "'")
		}
		dup[repos.Path] = true
	}

	// Validate if duplicate profiles[]/name exist
	dup = make(map[string]bool, len(lockJSON.Profiles))
	for i := range lockJSON.Profiles {
		profile := &lockJSON.Profiles[i]
		if _, exists := dup[profile.Name]; exists {
			return errors.New("duplicate profile '" + profile.Name + "'")
		}
		dup[profile.Name] = true
	}

	// Validate if duplicate profiles[]/repos_path[] exist
	for i := range lockJSON.Profiles {
		profile := &lockJSON.Profiles[i]
		dup = make(map[string]bool, len(lockJSON.Profiles)*10)
		for _, reposPath := range profile.ReposPath {
			if _, exists := dup[reposPath]; exists {
				return errors.New("duplicate '" + reposPath + "' (repos_path) in profile '" + profile.Name + "'")
			}
			dup[reposPath] = true
		}
	}

	// Validate if active_profile exists in profiles[]/name
	found := false
	for i := range lockJSON.Profiles {
		profile := &lockJSON.Profiles[i]
		if profile.Name == lockJSON.CurrentProfileName {
			found = true
			break
		}
	}
	if !found {
		return errors.New("'" + lockJSON.CurrentProfileName + "' (active_profile) doesn't exist in profiles")
	}

	// Validate if profiles[]/repos_path[] exists in repos[]/path
	for i := range lockJSON.Profiles {
		profile := &lockJSON.Profiles[i]
		for j, reposPath := range profile.ReposPath {
			found := false
			for k := range lockJSON.Repos {
				if reposPath == lockJSON.Repos[k].Path {
					found = true
					break
				}
			}
			if !found {
				return errors.New(
					"'" + reposPath + "' (profiles[" + strconv.Itoa(i) +
						"].repos_path[" + strconv.Itoa(j) + "]) doesn't exist in repos")
			}
		}
	}

	// Validate if trx_id is equal or greater than repos[]/trx_id
	index := -1
	var max int64
	for i := range lockJSON.Repos {
		repos := &lockJSON.Repos[i]
		if max < repos.TrxID {
			index = i
			max = repos.TrxID
		}
	}
	if max > lockJSON.TrxID {
		return errors.New("'" + strconv.FormatInt(max, 10) + "' (repos[" + strconv.Itoa(index) + "].trx_id) " +
			"is greater than '" + strconv.FormatInt(lockJSON.TrxID, 10) + "' (trx_id)")
	}

	return nil
}

func validateMissing(lockJSON *LockJSON) error {
	if lockJSON.Version == 0 {
		return errors.New("missing: version")
	}
	if lockJSON.TrxID == 0 {
		return errors.New("missing: trx_id")
	}
	if lockJSON.Repos == nil {
		return errors.New("missing: repos")
	}
	for i := range lockJSON.Repos {
		repos := &lockJSON.Repos[i]
		if repos.Type == "" {
			return errors.New("missing: repos[" + strconv.Itoa(i) + "].type")
		}
		switch repos.Type {
		case ReposGitType:
			if repos.Version == "" {
				return errors.New("missing: repos[" + strconv.Itoa(i) + "].version")
			}
			fallthrough
		case ReposStaticType:
			if repos.TrxID == 0 {
				return errors.New("missing: repos[" + strconv.Itoa(i) + "].trx_id")
			}
			if repos.Path == "" {
				return errors.New("missing: repos[" + strconv.Itoa(i) + "].path")
			}
		default:
			return errors.New("repos[" + strconv.Itoa(i) + "].type is invalid type: " + string(repos.Type))
		}
	}
	if lockJSON.Profiles == nil {
		return errors.New("missing: profiles")
	}
	for i := range lockJSON.Profiles {
		profile := &lockJSON.Profiles[i]
		if profile.Name == "" {
			return errors.New("missing: profile[" + strconv.Itoa(i) + "].name")
		}
		if profile.ReposPath == nil {
			return errors.New("missing: profile[" + strconv.Itoa(i) + "].repos_path")
		}
		for j, reposPath := range profile.ReposPath {
			if reposPath == "" {
				return errors.New("missing: profile[" + strconv.Itoa(i) + "].repos_path[" + strconv.Itoa(j) + "]")
			}
		}
	}
	return nil
}

func (lockJSON *LockJSON) Write() error {
	// Validate lock.json
	err := validate(lockJSON)
	if err != nil {
		return err
	}

	// Mkdir all if lock.json's directory does not exist
	lockfile := pathutil.LockJSON()
	if !pathutil.Exists(filepath.Dir(lockfile)) {
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

func (profs *ProfileList) FindByName(name string) (*Profile, error) {
	for i := range *profs {
		if (*profs)[i].Name == name {
			return &(*profs)[i], nil
		}
	}
	return nil, errors.New("profile '" + name + "' does not exist")
}

func (profs *ProfileList) FindIndexByName(name string) int {
	for i := range *profs {
		if (*profs)[i].Name == name {
			return i
		}
	}
	return -1
}

func (profs *ProfileList) RemoveAllReposPath(reposPath string) error {
	for i := range *profs {
		for j := range (*profs)[i].ReposPath {
			if (*profs)[i].ReposPath[j] == reposPath {
				(*profs)[i].ReposPath = append(
					(*profs)[i].ReposPath[:j],
					(*profs)[i].ReposPath[j+1:]...,
				)
				return nil
			}
		}
	}
	return errors.New("no matching profiles[]/repos_path[]: " + reposPath)
}

func (reposList *ReposList) Contains(reposPath string) bool {
	_, err := reposList.FindByPath(reposPath)
	return err == nil
}

func (reposList *ReposList) FindByPath(reposPath string) (*Repos, error) {
	for i := range *reposList {
		repos := &(*reposList)[i]
		if repos.Path == reposPath {
			return repos, nil
		}
	}
	return nil, errors.New("repos '" + reposPath + "' does not exist")
}

func (reposList *ReposList) RemoveAllByPath(reposPath string) error {
	for i := range *reposList {
		if (*reposList)[i].Path == reposPath {
			*reposList = append((*reposList)[:i], (*reposList)[i+1:]...)
			return nil
		}
	}
	return errors.New("no matching repos[]/path: " + reposPath)
}

func (reposPathList *profReposPath) Contains(reposPath string) bool {
	return reposPathList.IndexOf(reposPath) >= 0
}

func (reposPathList *profReposPath) IndexOf(reposPath string) int {
	for i := range *reposPathList {
		if (*reposPathList)[i] == reposPath {
			return i
		}
	}
	return -1
}

func (lockJSON *LockJSON) GetReposListByProfile(profile *Profile) ([]Repos, error) {
	var reposList []Repos
	for _, reposPath := range profile.ReposPath {
		repos, err := lockJSON.Repos.FindByPath(reposPath)
		if err != nil {
			return nil, err
		}
		reposList = append(reposList, *repos)
	}
	return reposList, nil
}
