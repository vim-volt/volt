package builder

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/src-d/go-git.v4"

	"github.com/vim-volt/volt/cmd/buildinfo"
	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/pathutil"
	"github.com/vim-volt/volt/plugconf"
)

type symlinkBuilder struct {
	baseBuilder
}

// TODO: rollback when return err (!= nil)
func (builder *symlinkBuilder) Build(buildInfo *buildinfo.BuildInfo, buildReposMap map[string]*buildinfo.Repos) error {
	// Exit if vim executable was not found in PATH
	if _, err := pathutil.VimExecutable(); err != nil {
		return err
	}

	// Remove vim volt dir every times
	vimVoltDir := pathutil.VimVoltDir()
	os.RemoveAll(vimVoltDir)
	if pathutil.Exists(vimVoltDir) {
		return errors.New("failed to remove " + vimVoltDir)
	}

	// Get current profile's repos list
	lockJSON, err := lockjson.Read()
	if err != nil {
		return errors.New("could not read lock.json: " + err.Error())
	}
	profile, reposList, err := builder.getCurrentProfileAndReposList(lockJSON)
	if err != nil {
		return err
	}

	logger.Info("Installing vimrc and gvimrc ...")

	vimDir := pathutil.VimDir()
	vimrcPath := filepath.Join(vimDir, pathutil.Vimrc)
	gvimrcPath := filepath.Join(vimDir, pathutil.Gvimrc)
	err = builder.installVimrcAndGvimrc(
		lockJSON.CurrentProfileName, vimrcPath, gvimrcPath, profile.UseVimrc, profile.UseGvimrc,
	)
	if err != nil {
		return err
	}

	// Mkdir opt dir
	optDir := pathutil.VimVoltOptDir()
	os.MkdirAll(optDir, 0755)
	if !pathutil.Exists(optDir) {
		return errors.New("could not create " + optDir)
	}

	buildInfo.Repos = make([]buildinfo.Repos, 0, len(reposList))
	copyBuilder := &copyBuilder{}
	for i := range reposList {
		src := pathutil.FullReposPathOf(reposList[i].Path)
		dst := pathutil.PackReposPathOf(reposList[i].Path)
		// Open a repository to determine it is bare repository or not
		r, err := git.PlainOpen(src)
		if err != nil {
			return errors.New("failed to open repository: " + err.Error())
		}
		cfg, err := r.Config()
		if err != nil {
			return errors.New("failed to get repository config: " + err.Error())
		}
		if reposList[i].Type == lockjson.ReposGitType && cfg.Core.IsBare {
			// * Copy files from git objects under vim dir
			// * Run ":helptags" to generate tags file
			done := make(chan actionReposResult)
			copyBuilder.updateBareGitRepos(r, src, dst, &reposList[i], done)
			result := <-done
			if result.err != nil {
				return result.err
			}
		} else {
			// Make symlinks under vim dir
			if err := os.Symlink(src, dst); err != nil {
				return err
			}
			// Run ":helptags" to generate tags file
			if err = builder.helptags(reposList[i].Path); err != nil {
				return err
			}
		}
		logger.Debug("Installing " + string(reposList[i].Type) + " repository " + reposList[i].Path + " ... Done.")
		// Make build-info.json data
		buildInfo.Repos = append(buildInfo.Repos, buildinfo.Repos{
			Type:    reposList[i].Type,
			Path:    reposList[i].Path,
			Version: reposList[i].Version,
		})
	}

	// Write bundled plugconf file
	content, merr := plugconf.GenerateBundlePlugconf(reposList)
	if merr.ErrorOrNil() != nil {
		// Return vim script parse errors
		return merr
	}
	os.MkdirAll(filepath.Dir(pathutil.BundledPlugConf()), 0755)
	err = ioutil.WriteFile(pathutil.BundledPlugConf(), content, 0644)
	if err != nil {
		return err
	}

	// Write build-info.json
	return buildInfo.Write()
}
