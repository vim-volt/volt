package lockjson

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/vim-volt/go-volt/pathutil"
)

type LockJson struct {
	ActiveProfile string    `json:"active_profile"`
	LoadInit      bool      `json:"load_init"`
	Repos         []Repos   `json:"repos"`
	Profiles      []Profile `json:"profiles"`
}

type Repos struct {
	Path    string `json:"path"`
	Version string `json:"version"`
	Active  bool   `json:"active"`
}

type Profile struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	LoadInit bool   `json:"load_init"`
}

func InitialLockJson() *LockJson {
	return &LockJson{
		ActiveProfile: "default",
		LoadInit:      true,
		Repos:         make([]Repos, 0),
		Profiles:      make([]Profile, 0),
	}
}

func Read() (*LockJson, error) {
	// Return initial lock.json struct if lockfile does not exist
	lockfile := pathutil.LockJson()
	if _, err := os.Stat(lockfile); os.IsNotExist(err) {
		return InitialLockJson(), nil
	}

	// Read lock.json and return it
	bytes, err := ioutil.ReadFile(lockfile)
	if err != nil {
		return nil, err
	}
	var lockJSON LockJson
	err = json.Unmarshal(bytes, &lockJSON)
	if err != nil {
		return nil, err
	}
	return &lockJSON, nil
}

func Write(lockJSON *LockJson) error {
	// Mkdir all if lock.json's directory does not exist
	lockfile := pathutil.LockJson()
	if _, err := os.Stat(filepath.Dir(lockfile)); os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(lockfile), 0755)
		if err != nil {
			return err
		}
	}

	// Write to lock.json
	bytes, err := json.Marshal(lockJSON)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(pathutil.LockJson(), bytes, 0644)
}
