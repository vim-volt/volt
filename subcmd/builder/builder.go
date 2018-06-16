package builder

import (
	"errors"
	"os"

	"github.com/vim-volt/volt/buildinfo"
	"github.com/vim-volt/volt/config"
	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/pathutil"
)

// Builder creates/updates ~/.vim/pack/volt directory
type Builder interface {
	Build(buildInfo *buildinfo.BuildInfo, buildReposMap map[pathutil.ReposPath]*buildinfo.Repos) error
}

const currentBuildInfoVersion = 2

// Build creates/updates ~/.vim/pack/volt directory
func Build(full bool, lockJSON *lockjson.LockJSON, cfg *config.Config) error {
	// Get builder
	blder, err := getBuilder(cfg.Build.Strategy, lockJSON)
	if err != nil {
		return err
	}

	// Read ~/.vim/pack/volt/opt/build-info.json
	buildInfo, err := buildinfo.Read()
	if err != nil {
		return err
	}

	// Do full build when:
	// * build-info.json's version is different with current version
	// * build-info.json's strategy is different with config
	// * config strategy is symlink
	if buildInfo.Version != currentBuildInfoVersion ||
		buildInfo.Strategy != cfg.Build.Strategy ||
		cfg.Build.Strategy == config.SymlinkBuilder {
		full = true
	}
	buildInfo.Version = currentBuildInfoVersion
	buildInfo.Strategy = cfg.Build.Strategy

	// Put repos into map to be able to search with O(1).
	// Use empty build-info.json map if the -full option was given
	// because the repos info is unnecessary because it is not referenced.
	var buildReposMap map[pathutil.ReposPath]*buildinfo.Repos
	optDir := pathutil.VimVoltOptDir()
	if full {
		buildReposMap = make(map[pathutil.ReposPath]*buildinfo.Repos)
		logger.Info("Full building " + optDir + " directory ...")
	} else {
		buildReposMap = make(map[pathutil.ReposPath]*buildinfo.Repos, len(buildInfo.Repos))
		for i := range buildInfo.Repos {
			repos := &buildInfo.Repos[i]
			buildReposMap[repos.Path] = repos
		}
		logger.Info("Building " + optDir + " directory ...")
	}

	// Remove ~/.vim/pack/volt/ if -full option was given
	if full {
		vimVoltDir := pathutil.VimVoltDir()
		os.RemoveAll(vimVoltDir)
		if pathutil.Exists(vimVoltDir) {
			return errors.New("failed to remove " + vimVoltDir)
		}
	}

	return blder.Build(buildInfo, buildReposMap)
}

func getBuilder(strategy string, lockJSON *lockjson.LockJSON) (Builder, error) {
	base := &BaseBuilder{lockJSON: lockJSON}
	switch strategy {
	case config.SymlinkBuilder:
		return &symlinkBuilder{base}, nil
	case config.CopyBuilder:
		return &copyBuilder{base}, nil
	default:
		return nil, errors.New("unknown builder type: " + strategy)
	}
}
