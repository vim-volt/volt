package builder

import (
	"errors"

	"github.com/vim-volt/volt/cmd/buildinfo"
	"github.com/vim-volt/volt/config"
)

type Builder interface {
	Build(buildInfo *buildinfo.BuildInfo, buildReposMap map[string]*buildinfo.Repos) error
}

func Get(strategy string) (Builder, error) {
	switch strategy {
	case config.SymlinkBuilder:
		return &symlinkBuilder{}, nil
	case config.CopyBuilder:
		return &copyBuilder{}, nil
	default:
		return nil, errors.New("unknown builder type: " + strategy)
	}
}
