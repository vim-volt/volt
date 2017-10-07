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

type rebuildFlagsType struct {
	helped bool
	full   bool
}

var rebuildFlags rebuildFlagsType

func init() {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Println(`
Usage
  volt rebuild [-help] [-full]

Quick example
  $ volt rebuild        # rebuilds directories under ~/.vim/pack/volt
  $ volt rebuild -full  # full rebuild (remove ~/.vim/pack/volt, and re-create all)

Description
  Rebuild ~/.vim/pack/volt/start/ and ~/.vim/pack/volt/opt/ directory:
    1. Copy repositories' files into ~/.vim/pack/volt/start/ and ~/.vim/pack/volt/opt/
      * If the repository is git repository, extract files from locked revision of tree object and copy them into above vim directories
      * If the repository is static repository (imported non-git directory by "volt add" command), copy files into above vim directories
    2. Remove directories from above vim directories, which exist in ~/.vim/pack/volt/build-info.json but not in $VOLTPATH/lock.json

  ~/.vim/pack/volt/build-info.json is a file which holds the information that what vim plugins are installed in ~/.vim/pack/volt/ and its type (git repository, static repository, or system repository), its version. A user normally doesn't need to know the contents of build-info.json .

  If -full option was given, remove all directories in ~/.vim/pack/volt/start/ and ~/.vim/pack/volt/opt/ , and copy repositories' files into above vim directories.
  Otherwise, it will perform smart rebuild: copy / remove only changed repositories' files.
`)
		fmt.Println("Options")
		fs.PrintDefaults()
		fmt.Println()
		rebuildFlags.helped = true
	}
	fs.BoolVar(&rebuildFlags.full, "full", false, "full rebuild")

	cmdFlagSet["rebuild"] = fs
}

type rebuildCmd struct{}

func Rebuild(args []string) int {
	cmd := rebuildCmd{}

	// Parse args
	fs := cmdFlagSet["rebuild"]
	fs.Parse(args)
	if rebuildFlags.helped {
		return 0
	}

	// Begin transaction
	err := transaction.Create()
	if err != nil {
		logger.Error("Failed to begin transaction:", err.Error())
		return 11
	}
	defer transaction.Remove()

	err = cmd.doRebuild(rebuildFlags.full)
	if err != nil {
		logger.Error("Failed to rebuild:", err.Error())
		return 12
	}

	return 0
}

const currentRebuildVersion = 1

type buildInfoType struct {
	Repos   reposList `json:"repos"`
	Version int64     `json:"version"`
}

type reposList []repos

type repos struct {
	Type    reposType `json:"type"`
	Path    string    `json:"path"`
	Version string    `json:"version"`
	Files   []file    `json:"files,omitempty"`
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

func (reposList *reposList) removeAllByReposPath(reposPath string) {
	for i := range *reposList {
		repos := &(*reposList)[i]
		if repos.Path == reposPath {
			*reposList = append((*reposList)[:i], (*reposList)[i+1:]...)
			break
		}
	}
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
	profile, reposList, err := cmd.getActiveProfileAndReposList(lockJSON)
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

	// Read ~/.vim/pack/volt/start/build-info.json
	buildInfo, err := cmd.readBuildInfo()
	if err != nil {
		return err
	}

	// Do -full rebuild when build-info.json's version is different
	if buildInfo.Version != currentRebuildVersion {
		full = true
	}
	buildInfo.Version = currentRebuildVersion

	// Put repos into map to be able to search with O(1).
	// Use empty build-info.json map if the -full option was given
	// because the repos info is unnecessary because it is not referenced.
	var buildReposMap map[string]*repos
	if full {
		buildReposMap = make(map[string]*repos)
		logger.Info("Full rebuilding " + startDir + " directory ...")
	} else {
		buildReposMap = make(map[string]*repos, len(buildInfo.Repos))
		for i := range buildInfo.Repos {
			repos := &buildInfo.Repos[i]
			buildReposMap[repos.Path] = repos
		}
		logger.Info("Rebuilding " + startDir + " directory ...")
	}

	// Remove ~/.vim/pack/volt/ if -full option was given
	if full {
		vimVoltDir := pathutil.VimVoltDir()
		err = os.RemoveAll(vimVoltDir)
		if err != nil {
			return errors.New("failed to remove " + vimVoltDir + ": " + err.Error())
		}
	}

	logger.Info("Installing vimrc and gvimrc ...")

	// Install vimrc
	err = cmd.installRCFile(
		lockJSON.ActiveProfile,
		"vimrc.vim",
		filepath.Join(vimDir, "vimrc"),
		profile.UseVimrc,
	)
	if err != nil {
		return err
	}

	// Install gvimrc
	err = cmd.installRCFile(
		lockJSON.ActiveProfile,
		"gvimrc.vim",
		filepath.Join(vimDir, "gvimrc"),
		profile.UseGvimrc,
	)
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
	copyDone := make(chan actionReposResult, len(reposList))
	copyCount := 0
	for i := range reposList {
		if reposList[i].Type == lockjson.ReposGitType {
			buildRepos, exists := buildReposMap[reposList[i].Path]
			if !exists ||
				!pathutil.Exists(pathutil.FullReposPathOf(reposList[i].Path)) ||
				cmd.hasChangedGitRepos(&reposList[i], buildRepos) {
				go cmd.updateGitRepos(&reposList[i], copyDone)
				copyCount++
			}
		} else if reposList[i].Type == lockjson.ReposStaticType {
			buildRepos, exists := buildReposMap[reposList[i].Path]
			if !exists ||
				!pathutil.Exists(pathutil.FullReposPathOf(reposList[i].Path)) ||
				cmd.hasChangedStaticRepos(&reposList[i], buildRepos, startDir) {
				go cmd.updateStaticRepos(&reposList[i], copyDone)
				copyCount++
			}
		} else {
			copyDone <- actionReposResult{
				errors.New("invalid repository type: " + string(reposList[i].Type)),
				&reposList[i],
			}
		}
	}

	// Remove all repositories found in build-info.json, but not in lock.json
	var removeList []repos
	for i := range buildInfo.Repos {
		if !lockJSON.Repos.Contains(buildInfo.Repos[i].Path) {
			removeList = append(removeList, buildInfo.Repos[i])
		}
	}
	removeDone := make(chan actionReposResult, len(removeList))
	for i := range removeList {
		go func(repos *repos) {
			err := os.RemoveAll(pathutil.FullReposPathOf(repos.Path))
			logger.Info("Removing " + string(repos.Type) + " repository " + repos.Path + " ... Done.")
			removeDone <- actionReposResult{
				err,
				&lockjson.Repos{Path: repos.Path},
			}
		}(&removeList[i])
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

	// Wait remove
	for i := 0; i < len(removeList); i++ {
		result := <-removeDone
		if result.err != nil {
			logger.Error("Failed to remove " + result.repos.Path + ": " + result.err.Error())
		} else {
			buildInfo.Repos.removeAllByReposPath(result.repos.Path)
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

func (cmd *rebuildCmd) installRCFile(profileName, srcRCFileName, dst string, install bool) error {
	// Return error if destination file has magic comment
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

	// Skip if use_vimrc/use_gvimrc is false or rc file does not exist
	src := pathutil.RCFileOf(profileName, srcRCFileName)
	if !install || !pathutil.Exists(src) {
		return nil
	}

	return cmd.copyFileWithMagicComment(src, dst)
}

const magicComment = "\" NOTE: this file was generated by volt. please modify original file.\n"
const magicCommentNext = "\" Original file: %s\n\n"

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

	_, err = writer.Write([]byte(fmt.Sprintf(magicCommentNext, src)))
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, reader)
	return err
}

type actionReposResult struct {
	err   error
	repos *lockjson.Repos
}

func (*rebuildCmd) getActiveProfileAndReposList(lockJSON *lockjson.LockJSON) (*lockjson.Profile, []lockjson.Repos, error) {
	// Find active profile
	profile, err := lockJSON.Profiles.FindByName(lockJSON.ActiveProfile)
	if err != nil {
		// this must not be occurred because lockjson.Read()
		// validates that the matching profile exists
		return nil, nil, err
	}

	reposList, err := lockJSON.GetReposListByProfile(profile)
	return profile, reposList, err
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
func (cmd *rebuildCmd) updateGitRepos(repos *lockjson.Repos, done chan actionReposResult) {
	src := pathutil.FullReposPathOf(repos.Path)
	dst := pathutil.PackReposPathOf(repos.Path)

	// Remove ~/.vim/volt/start/{repos}
	err := os.RemoveAll(dst)
	if err != nil {
		done <- actionReposResult{
			errors.New("failed to remove repository: " + err.Error()),
			repos,
		}
		return
	}

	// Open ~/volt/repos/{repos}
	r, err := git.PlainOpen(src)
	if err != nil {
		done <- actionReposResult{
			errors.New("failed to open repository: " + err.Error()),
			repos,
		}
		return
	}

	// Get locked commit hash
	commit := plumbing.NewHash(repos.Version)
	commitObj, err := r.CommitObject(commit)
	if err != nil {
		done <- actionReposResult{
			errors.New("failed to get HEAD commit object: " + err.Error()),
			repos,
		}
		return
	}

	// Get tree hash of commit hash
	tree, err := r.TreeObject(commitObj.TreeHash)
	if err != nil {
		done <- actionReposResult{
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
		os.MkdirAll(filepath.Dir(filename), 0755)
		ioutil.WriteFile(filename, []byte(contents), osMode)
		return nil
	})
	if err != nil {
		done <- actionReposResult{err, repos}
		return
	}

	done <- actionReposResult{nil, repos}
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
func (cmd *rebuildCmd) updateStaticRepos(repos *lockjson.Repos, done chan actionReposResult) {
	src := pathutil.FullReposPathOf(repos.Path)
	dst := pathutil.PackReposPathOf(repos.Path)

	// Remove ~/.vim/volt/start/{repos}
	err := os.RemoveAll(dst)
	if err != nil {
		done <- actionReposResult{
			errors.New("failed to remove repository: " + err.Error()),
			repos,
		}
		return
	}

	// Copy ~/volt/repos/{repos} to ~/.vim/volt/start/{repos}
	err = fileutil.CopyDir(src, dst)
	if err != nil {
		done <- actionReposResult{
			errors.New("failed to copy static directory: " + err.Error()),
			repos,
		}
		return
	}

	done <- actionReposResult{nil, repos}
}
