package buildinfo

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"

	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/pathutil"
)

type BuildInfo struct {
	Repos    ReposList `json:"repos"`
	Version  int64     `json:"version"`
	Strategy string    `json:"strategy"`
}

type ReposList []Repos

type Repos struct {
	Type          lockjson.ReposType `json:"type"`
	Path          pathutil.ReposPath `json:"path"`
	Version       string             `json:"version"`
	Files         FileMap            `json:"files,omitempty"`
	DirtyWorktree bool               `json:"dirty_worktree,omitempty"`
}

// key: filepath, value: version
type FileMap map[string]string

func Read() (*BuildInfo, error) {
	// Return initial build-info.json struct
	// if the file does not exist
	file := pathutil.BuildInfoJSON()
	if !pathutil.Exists(file) {
		return &BuildInfo{}, nil
	}

	// Read build-info.json
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var buildInfo BuildInfo
	err = json.Unmarshal(bytes, &buildInfo)
	if err != nil {
		return nil, err
	}

	// Validate build-info.json
	err = buildInfo.validate()
	if err != nil {
		return nil, errors.Wrap(err, "validation failed: build-info.json")
	}

	return &buildInfo, nil
}

func (buildInfo *BuildInfo) Write() error {
	// Validate build-info.json
	err := buildInfo.validate()
	if err != nil {
		return errors.Wrap(err, "validation failed: build-info.json")
	}

	// Write to build-info.json
	bytes, err := json.MarshalIndent(buildInfo, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(pathutil.BuildInfoJSON(), bytes, 0644)
}

func (buildInfo *BuildInfo) validate() error {
	// Validate if repos do not have duplicate repository
	dupRepos := make(map[pathutil.ReposPath]bool, len(buildInfo.Repos))
	for i := range buildInfo.Repos {
		r := &buildInfo.Repos[i]
		if _, exists := dupRepos[r.Path]; exists {
			return errors.New("duplicate repos: " + r.Path.String())
		}
		dupRepos[r.Path] = true
	}
	return nil
}

func (reposList *ReposList) FindByReposPath(reposPath pathutil.ReposPath) *Repos {
	for i := range *reposList {
		repos := &(*reposList)[i]
		if repos.Path == reposPath {
			return repos
		}
	}
	return nil
}

func (reposList *ReposList) RemoveByReposPath(reposPath pathutil.ReposPath) {
	for i := range *reposList {
		repos := &(*reposList)[i]
		if repos.Path == reposPath {
			*reposList = append((*reposList)[:i], (*reposList)[i+1:]...)
			break
		}
	}
}
