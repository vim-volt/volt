package lockjson

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

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

	return &lockJSON, nil
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
