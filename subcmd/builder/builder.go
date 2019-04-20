package builder

import (
	"github.com/pkg/errors"
	"os"

	"github.com/vim-volt/volt/config"
	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/pathutil"
	"github.com/vim-volt/volt/subcmd/buildinfo"
)

// Builder creates/updates ~/.vim/pack/volt directory
type Builder interface {
	Build(buildInfo *buildinfo.BuildInfo, buildReposMap map[pathutil.ReposPath]*buildinfo.Repos) error
}

const currentBuildInfoVersion = 2

// Build creates/updates ~/.vim/pack/volt directory
func Build(full bool) error {
	// Read config.toml
	cfg, err := config.Read()
	if err != nil {
		return errors.Wrap(err, "could not read config.toml")
	}

	// Get builder
	blder, err := getBuilder(cfg.Build.Strategy)
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

func getBuilder(strategy string) (Builder, error) {
	switch strategy {
	case config.SymlinkBuilder:
		return &symlinkBuilder{}, nil
	case config.CopyBuilder:
		return &copyBuilder{}, nil
	default:
		return nil, errors.New("unknown builder type: " + strategy)
	}
}
