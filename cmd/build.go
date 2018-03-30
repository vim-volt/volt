package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/vim-volt/volt/cmd/builder"
	"github.com/vim-volt/volt/cmd/buildinfo"
	"github.com/vim-volt/volt/config"
	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/pathutil"
	"github.com/vim-volt/volt/transaction"
)

func init() {
	cmdMap["build"] = &buildCmd{}
}

type buildCmd struct {
	helped bool
	full   bool
}

func (cmd *buildCmd) ProhibitRootExecution(args []string) bool { return true }

func (cmd *buildCmd) FlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Print(`
Usage
  volt build [-help] [-full]

Quick example
  $ volt build        # builds directories under ~/.vim/pack/volt
  $ volt build -full  # full build (remove ~/.vim/pack/volt, and re-create all)

Description
  Build ~/.vim/pack/volt/opt/ directory:
    1. Copy repositories' files into ~/.vim/pack/volt/opt/
      * If the repository is git repository, extract files from locked revision of tree object and copy them into above vim directories
      * If the repository is static repository (imported non-git directory by "volt add" command), copy files into above vim directories
    2. Remove directories from above vim directories, which exist in ~/.vim/pack/volt/build-info.json but not in $VOLTPATH/lock.json

  ~/.vim/pack/volt/build-info.json is a file which holds the information that what vim plugins are installed in ~/.vim/pack/volt/ and its type (git repository, static repository, or system repository), its version. A user normally doesn't need to know the contents of build-info.json .

  If -full option was given, remove all directories in ~/.vim/pack/volt/opt/ , and copy repositories' files into above vim directories.
  Otherwise, it will perform smart build: copy / remove only changed repositories' files.` + "\n\n")
		fmt.Println("Options")
		fs.PrintDefaults()
		fmt.Println()
		cmd.helped = true
	}
	fs.BoolVar(&cmd.full, "full", false, "full build")
	return fs
}

func (cmd *buildCmd) Run(args []string) int {
	// Parse args
	fs := cmd.FlagSet()
	fs.Parse(args)
	if cmd.helped {
		return 0
	}

	// Begin transaction
	err := transaction.Create()
	if err != nil {
		logger.Error("Failed to begin transaction:", err.Error())
		return 11
	}
	defer transaction.Remove()

	err = cmd.doBuild(cmd.full)
	if err != nil {
		logger.Error("Failed to build:", err.Error())
		return 12
	}

	return 0
}

const currentBuildInfoVersion = 2

func (cmd *buildCmd) doBuild(full bool) error {
	// Read config.toml
	cfg, err := config.Read()
	if err != nil {
		return errors.New("could not read config.toml: " + err.Error())
	}

	// Get builder
	builder, err := builder.Get(cfg.Build.Strategy)
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

	return builder.Build(buildInfo, buildReposMap)
}
