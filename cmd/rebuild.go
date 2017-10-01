package cmd

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vim-volt/volt/fileutil"
	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/pathutil"
	"github.com/vim-volt/volt/transaction"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type rebuildCmd struct{}

type rebuildFlags struct {
	full bool
}

func Rebuild(args []string) int {
	cmd := rebuildCmd{}

	// Parse args
	flags, err := cmd.parseArgs(args)
	if err != nil {
		logger.Error("Failed to parse args: " + err.Error())
		return 10
	}

	// Begin transaction
	err = transaction.Create()
	if err != nil {
		logger.Error("Failed to begin transaction:", err.Error())
		return 11
	}
	defer transaction.Remove()

	err = cmd.doRebuild(flags.full)
	if err != nil {
		logger.Error("Failed to rebuild:", err.Error())
		return 12
	}

	return 0
}

func (rebuildCmd) parseArgs(args []string) (*rebuildFlags, error) {
	var flags rebuildFlags
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Println(`
Usage
  volt rebuild [-help] [-full]

Description
  Rebuild ~/.vim/pack/volt/ directory.
  If -full was given, remove and update all repositories again.
  If -full was not given, remove and update only updated repositories.

Options`)
		fs.PrintDefaults()
		fmt.Println()
	}
	fs.BoolVar(&flags.full, "full", false, "full rebuild")
	fs.Parse(args)

	return &flags, nil
}

type buildInfoType struct {
	Repos reposList `json:"repos"`
}

type reposList []repos

type repos struct {
	Type    reposType `json:"type"`
	Path    string    `json:"path"`
	Version string    `json:"version"`
	Files   []file    `json:"files"`
}

type reposType string

const (
	reposGitType    reposType = "git"
	reposStaticType reposType = "static"
	reposSystemType reposType = "system"
)

type file struct {
	Path    string `json:"path"`
	Version string `json:"version"`
}

func (cmd *rebuildCmd) readBuildInfo() (*buildInfoType, error) {
	// Return initial build-info.json struct
	// if the file does not exist
	file := pathutil.BuildInfoJSON()
	if !pathutil.Exists(file) {
		return &buildInfoType{}, nil
	}

	// Read build-info.json
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var buildInfo buildInfoType
	err = json.Unmarshal(bytes, &buildInfo)
	if err != nil {
		return nil, err
	}

	// Validate build-info.json
	err = buildInfo.validate()
	if err != nil {
		return nil, errors.New("validation failed: build-info.json: " + err.Error())
	}

	return &buildInfo, nil
}

func (buildInfo *buildInfoType) write() error {
	// Validate build-info.json
	err := buildInfo.validate()
	if err != nil {
		return errors.New("validation failed: build-info.json: " + err.Error())
	}

	// Write to build-info.json
	bytes, err := json.MarshalIndent(buildInfo, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(pathutil.BuildInfoJSON(), bytes, 0644)
}

func (buildInfo *buildInfoType) validate() error {
	// Validate if repos do not have duplicate repository
	dupRepos := make(map[string]bool, len(buildInfo.Repos))
	for i := range buildInfo.Repos {
		r := &buildInfo.Repos[i]
		if _, exists := dupRepos[r.Path]; exists {
			return errors.New("duplicate repos: " + r.Path)
		}
		dupRepos[r.Path] = true

		// Validate if files do not have duplicate repository
		dupFiles := make(map[string]bool, len(r.Files))
		for j := range r.Files {
			f := &r.Files[j]
			if _, exists := dupFiles[f.Path]; exists {
				return errors.New(r.Path + ": duplicate files: " + f.Path)
			}
			dupFiles[f.Path] = true
		}
	}
	return nil
}

func (reposList *reposList) findByReposPath(reposPath string) *repos {
	for i := range *reposList {
		repos := &(*reposList)[i]
		if repos.Path == reposPath {
			return repos
		}
	}
	return nil
}

func (cmd *rebuildCmd) doRebuild(full bool) error {
	vimDir := pathutil.VimDir()
	startDir := pathutil.VimVoltStartDir()

	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		return errors.New("could not read lock.json: " + err.Error())
	}

	// Get active profile's repos list
	reposList, err := cmd.getActiveProfileRepos(lockJSON)
	if err != nil {
		return err
	}

	// Exit with an error if vimrc or gvimrc without magic comment exists
	for _, file := range pathutil.LookUpVimrcOrGvimrc() {
		err = cmd.shouldHaveMagicComment(file)
		// If the file does not have magic comment
		if err != nil {
			return errors.New("already exists user vimrc or gvimrc: " + err.Error())
		}
	}

	var buildInfo *buildInfoType
	if full {
		// Use empty build-info.json struct
		// if the -full option was given
		buildInfo = &buildInfoType{}
	} else {
		// Read ~/.vim/pack/volt/start/build-info.json
		var err error
		buildInfo, err = cmd.readBuildInfo()
		if err != nil {
			return err
		}
	}

	// Put repos into map to be able to search with O(1)
	buildReposMap := make(map[string]*repos, len(buildInfo.Repos))
	for i := range buildInfo.Repos {
		repos := &buildInfo.Repos[i]
		buildReposMap[repos.Path] = repos
	}

	logger.Info("Rebuilding " + startDir + " directory ...")
	logger.Info("Installing vimrc and gvimrc ...")

	// Install vimrc and gvimrc
	err = cmd.installRCFile(lockJSON.ActiveProfile, "vimrc.vim", filepath.Join(vimDir, "vimrc"))
	if err != nil {
		return err
	}
	err = cmd.installRCFile(lockJSON.ActiveProfile, "gvimrc.vim", filepath.Join(vimDir, "gvimrc"))
	if err != nil {
		return err
	}

	// Mkdir start dir
	os.MkdirAll(startDir, 0755)
	if !pathutil.Exists(startDir) {
		return errors.New("could not create " + startDir)
	}

	logger.Info("Installing all repositories files ...")

	// Copy all repositories files to startDir
	copyDone := make(chan copyReposResult, len(reposList))
	copyCount := 0
	for i := range reposList {
		if reposList[i].Type == lockjson.ReposGitType {
			buildRepos, exists := buildReposMap[reposList[i].Path]
			if !exists || cmd.hasChangedGitRepos(&reposList[i], buildRepos) {
				go cmd.updateGitRepos(&reposList[i], startDir, copyDone)
				copyCount++
			}
		} else if reposList[i].Type == lockjson.ReposStaticType {
			buildRepos, exists := buildReposMap[reposList[i].Path]
			if !exists || cmd.hasChangedStaticRepos(&reposList[i], buildRepos, startDir) {
				go cmd.updateStaticRepos(&reposList[i], startDir, copyDone)
				copyCount++
			}
		} else {
			copyDone <- copyReposResult{
				errors.New("invalid repository type: " + string(reposList[i].Type)),
				&reposList[i],
			}
		}
	}

	// Wait copy. construct buildInfo from the results
	modified := false
	for i := 0; i < copyCount; i++ {
		result := <-copyDone
		if result.err != nil {
			return errors.New("failed to copy repository '" + result.repos.Path + "': " + result.err.Error())
		} else if result.repos.Type == lockjson.ReposGitType {
			logger.Info("Installing git repository " + result.repos.Path + " ... Done.")
			r := buildInfo.Repos.findByReposPath(result.repos.Path)
			if r != nil {
				r.Version = result.repos.Version
			} else {
				buildInfo.Repos = append(
					buildInfo.Repos,
					repos{
						Type:    reposGitType,
						Path:    result.repos.Path,
						Version: result.repos.Version,
					},
				)
			}
			modified = true
		} else if result.repos.Type == lockjson.ReposStaticType {
			logger.Info("Installing static directory " + result.repos.Path + " ... Done.")
			r := buildInfo.Repos.findByReposPath(result.repos.Path)
			if r != nil {
				r.Version = time.Now().Format(time.RFC3339)
			} else {
				buildInfo.Repos = append(
					buildInfo.Repos,
					repos{
						Type:    reposStaticType,
						Path:    result.repos.Path,
						Version: time.Now().Format(time.RFC3339),
					},
				)
			}
			modified = true
		}
	}

	// Write to build-info.json if modified
	if modified {
		err = buildInfo.write()
		if err != nil {
			return err
		}
		logger.Info("Written build-info.json")
	}

	return nil
}

func (cmd *rebuildCmd) installRCFile(profileName, srcRCFileName, dst string) error {
	if pathutil.Exists(dst) {
		err := cmd.shouldHaveMagicComment(dst)
		// If the file does not have magic comment
		if err != nil {
			return err
		}
	}

	// Remove destination (~/.vim/vimrc or ~/.vim/gvimrc)
	os.Remove(dst)
	if pathutil.Exists(dst) {
		return errors.New("failed to remove " + dst)
	}

	// Skip if rc file does not exist
	src := pathutil.RCFileOf(profileName, srcRCFileName)
	if !pathutil.Exists(src) {
		return nil
	}

	return cmd.copyFileWithMagicComment(src, dst)
}

const magicComment = "\" NOTE: this file was generated by volt. please modify original file.\n"

// Return error if the magic comment does not exist
func (*rebuildCmd) shouldHaveMagicComment(dst string) error {
	reader, err := os.Open(dst)
	if err != nil {
		return err
	}
	defer reader.Close()

	magic := []byte(magicComment)
	read := make([]byte, len(magic))
	n, err := reader.Read(read)
	if err != nil || n < len(magicComment) {
		return errors.New("'" + dst + "' does not have magic comment")
	}

	for i := range magic {
		if magic[i] != read[i] {
			return errors.New("'" + dst + "' does not have magic comment")
		}
	}
	return nil
}

func (*rebuildCmd) copyFileWithMagicComment(src, dst string) error {
	reader, err := os.Open(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	writer, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = writer.Write([]byte(magicComment))
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, reader)
	return err
}

type copyReposResult struct {
	err   error
	repos *lockjson.Repos
}

func (*rebuildCmd) getActiveProfileRepos(lockJSON *lockjson.LockJSON) ([]lockjson.Repos, error) {
	// Find active profile
	profile, err := lockJSON.Profiles.FindByName(lockJSON.ActiveProfile)
	if err != nil {
		// this must not be occurred because lockjson.Read()
		// validates that the matching profile exists
		return nil, err
	}

	return lockJSON.GetReposListByProfile(profile)
}

func (*rebuildCmd) getLatestModTime(path string) (time.Time, error) {
	mtime := time.Unix(0, 0)
	err := fileutil.Traverse(path, func(fi os.FileInfo) {
		t := fi.ModTime()
		if mtime.Before(t) {
			mtime = t
		}
	})
	if err != nil {
		return time.Now(), errors.New("failed to readdir: " + err.Error())
	}
	return mtime, nil
}

func (*rebuildCmd) hasChangedGitRepos(repos *lockjson.Repos, buildRepos *repos) bool {
	if repos.Version != buildRepos.Version {
		// repository has changed, do copy
		return true
	}
	return false
}

// Remove ~/.vim/volt/start/{repos} and copy from ~/volt/repos/{repos}
func (cmd *rebuildCmd) updateGitRepos(repos *lockjson.Repos, startDir string, done chan copyReposResult) {
	src := pathutil.FullReposPathOf(repos.Path)
	dst := filepath.Join(startDir, cmd.encodeReposPath(repos.Path))

	// Remove ~/.vim/volt/start/{repos}
	err := os.RemoveAll(dst)
	if err != nil {
		done <- copyReposResult{
			errors.New("failed to remove repository: " + err.Error()),
			repos,
		}
		return
	}

	// Open ~/volt/repos/{repos}
	r, err := git.PlainOpen(src)
	if err != nil {
		done <- copyReposResult{
			errors.New("failed to open repository: " + err.Error()),
			repos,
		}
		return
	}

	// Get locked commit hash
	commit := plumbing.NewHash(repos.Version)
	commitObj, err := r.CommitObject(commit)
	if err != nil {
		done <- copyReposResult{
			errors.New("failed to get HEAD commit object: " + err.Error()),
			repos,
		}
		return
	}

	// Get tree hash of commit hash
	tree, err := r.TreeObject(commitObj.TreeHash)
	if err != nil {
		done <- copyReposResult{
			errors.New("failed to get tree " + commit.String() + ": " + err.Error()),
			repos,
		}
		return
	}

	// Copy files
	err = tree.Files().ForEach(func(file *object.File) error {
		osMode, err := file.Mode.ToOSFileMode()
		if err != nil {
			return errors.New("failed to convert file mode: " + err.Error())
		}

		contents, err := file.Contents()
		if err != nil {
			return errors.New("failed get file contents: " + err.Error())
		}

		filename := filepath.Join(dst, file.Name)
		dir, _ := filepath.Split(filename)
		os.MkdirAll(dir, 0755)
		ioutil.WriteFile(filename, []byte(contents), osMode)
		return nil
	})
	if err != nil {
		done <- copyReposResult{err, repos}
		return
	}

	done <- copyReposResult{nil, repos}
}

func (*rebuildCmd) encodeReposPath(reposPath string) string {
	return strings.NewReplacer("_", "__", "/", "_").Replace(reposPath)
}

func (cmd *rebuildCmd) hasChangedStaticRepos(repos *lockjson.Repos, buildRepos *repos, startDir string) bool {
	src := pathutil.FullReposPathOf(repos.Path)

	// Get latest mtime of src
	srcModTime, err := cmd.getLatestModTime(src)
	if err != nil {
		// failed to readdir, do copy again
		return true
	}

	if buildRepos.Version == "" {
		// not found mtime, do copy again
		return true
	}

	// Get latest mtime of dst from build-info.json
	dstModTime, err := time.Parse(time.RFC3339, buildRepos.Version)
	if err != nil {
		// failed to parse datetime, do copy again
		return true
	}

	return dstModTime.Before(srcModTime)
}

// Remove ~/.vim/volt/start/{repos} and copy from ~/volt/repos/{repos}
func (cmd *rebuildCmd) updateStaticRepos(repos *lockjson.Repos, startDir string, done chan copyReposResult) {
	src := pathutil.FullReposPathOf(repos.Path)
	dst := filepath.Join(startDir, cmd.encodeReposPath(repos.Path))

	// Remove ~/.vim/volt/start/{repos}
	err := os.RemoveAll(dst)
	if err != nil {
		done <- copyReposResult{
			errors.New("failed to remove repository: " + err.Error()),
			repos,
		}
		return
	}

	// Copy ~/volt/repos/{repos} to ~/.vim/volt/start/{repos}
	err = fileutil.CopyDir(src, dst)
	if err != nil {
		done <- copyReposResult{
			errors.New("failed to copy static directory: " + err.Error()),
			repos,
		}
		return
	}

	done <- copyReposResult{nil, repos}
}
