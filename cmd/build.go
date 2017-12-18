package cmd

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/vim-volt/volt/fileutil"
	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/pathutil"
	"github.com/vim-volt/volt/plugconf"
	"github.com/vim-volt/volt/transaction"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type buildFlagsType struct {
	helped bool
	full   bool
}

var buildFlags buildFlagsType

var BuildModeInvalidType = os.ModeSymlink | os.ModeNamedPipe | os.ModeSocket | os.ModeDevice
var ErrBuildModeType = "does not allow symlink, named pipe, socket, device"

func init() {
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
		buildFlags.helped = true
	}
	fs.BoolVar(&buildFlags.full, "full", false, "full build")

	cmdFlagSet["build"] = fs
}

type buildCmd struct{}

func Build(args []string) int {
	cmd := buildCmd{}

	// Parse args
	fs := cmdFlagSet["build"]
	fs.Parse(args)
	if buildFlags.helped {
		return 0
	}

	// Begin transaction
	err := transaction.Create()
	if err != nil {
		logger.Error("Failed to begin transaction:", err.Error())
		return 11
	}
	defer transaction.Remove()

	err = cmd.doBuild(buildFlags.full)
	if err != nil {
		logger.Error("Failed to build:", err.Error())
		return 12
	}

	return 0
}

const currentBuildInfoVersion = 1

type buildInfoType struct {
	Repos   biReposList `json:"repos"`
	Version int64       `json:"version"`
}

type biReposList []biRepos

type biRepos struct {
	Type          reposType `json:"type"`
	Path          string    `json:"path"`
	Version       string    `json:"version"`
	Files         biFileMap `json:"files,omitempty"`
	DirtyWorktree bool      `json:"dirty_worktree,omitempty"`
}

type reposType string

const (
	reposGitType    reposType = "git"
	reposStaticType reposType = "static"
	reposSystemType reposType = "system"
)

// key: filepath, value: version
type biFileMap map[string]string

func (cmd *buildCmd) readBuildInfo() (*buildInfoType, error) {
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
	}
	return nil
}

func (reposList *biReposList) findByReposPath(reposPath string) *biRepos {
	for i := range *reposList {
		repos := &(*reposList)[i]
		if repos.Path == reposPath {
			return repos
		}
	}
	return nil
}

func (reposList *biReposList) removeByReposPath(reposPath string) {
	for i := range *reposList {
		repos := &(*reposList)[i]
		if repos.Path == reposPath {
			*reposList = append((*reposList)[:i], (*reposList)[i+1:]...)
			break
		}
	}
}

func (cmd *buildCmd) doBuild(full bool) error {
	// Exit if vim executable was not found in PATH
	if _, err := pathutil.VimExecutable(); err != nil {
		return err
	}

	vimDir := pathutil.VimDir()
	optDir := pathutil.VimVoltOptDir()

	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		return errors.New("could not read lock.json: " + err.Error())
	}

	// Get current profile's repos list
	profile, reposList, err := cmd.getCurrentProfileAndReposList(lockJSON)
	if err != nil {
		return err
	}

	// Check vimrc or gvimrc without magic comment exists
	rcFileExists := false
	for _, file := range pathutil.LookUpVimrcOrGvimrc() {
		err = cmd.shouldHaveMagicComment(file)
		// If the file does not have magic comment
		if err != nil {
			rcFileExists = true
		}
	}

	// Read ~/.vim/pack/volt/opt/build-info.json
	buildInfo, err := cmd.readBuildInfo()
	if err != nil {
		return err
	}

	// Do full build when build-info.json's version is different
	if buildInfo.Version != currentBuildInfoVersion {
		full = true
	}
	buildInfo.Version = currentBuildInfoVersion

	// Put repos into map to be able to search with O(1).
	// Use empty build-info.json map if the -full option was given
	// because the repos info is unnecessary because it is not referenced.
	var buildReposMap map[string]*biRepos
	if full {
		buildReposMap = make(map[string]*biRepos)
		logger.Info("Full building " + optDir + " directory ...")
	} else {
		buildReposMap = make(map[string]*biRepos, len(buildInfo.Repos))
		for i := range buildInfo.Repos {
			repos := &buildInfo.Repos[i]
			buildReposMap[repos.Path] = repos
		}
		logger.Info("Building " + optDir + " directory ...")
	}

	// Remove ~/.vim/pack/volt/ if -full option was given
	if full {
		vimVoltDir := pathutil.VimVoltDir()
		err = os.RemoveAll(vimVoltDir)
		if err != nil {
			return errors.New("failed to remove " + vimVoltDir + ": " + err.Error())
		}
	}

	if !rcFileExists {
		logger.Info("Installing vimrc and gvimrc ...")

		// Install vimrc
		err = cmd.installRCFile(
			lockJSON.CurrentProfileName,
			pathutil.ProfileVimrc,
			filepath.Join(vimDir, pathutil.Vimrc),
			profile.UseVimrc,
		)
		if err != nil {
			return err
		}

		// Install gvimrc
		err = cmd.installRCFile(
			lockJSON.CurrentProfileName,
			pathutil.ProfileGvimrc,
			filepath.Join(vimDir, pathutil.Gvimrc),
			profile.UseGvimrc,
		)
		if err != nil {
			return err
		}
	}

	// Mkdir opt dir
	os.MkdirAll(optDir, 0755)
	if !pathutil.Exists(optDir) {
		return errors.New("could not create " + optDir)
	}

	// Copy volt repos files to optDir
	copyDone, copyCount := cmd.copyReposList(buildReposMap, reposList, optDir)

	// Remove vim repos found in lock.json, but not in build-info.json
	removeDone, removeCount := cmd.removeReposList(buildInfo.Repos, lockJSON.Repos)

	// Wait copy
	var copyModified bool
	copyErr := cmd.waitCopyRepos(copyDone, copyCount, func(result *actionReposResult) error {
		logger.Info("Installing " + string(result.repos.Type) + " repository " + result.repos.Path + " ... Done.")
		// Construct buildInfo from the result
		cmd.constructBuildInfo(buildInfo, result)
		copyModified = true
		return nil
	})

	// Wait remove
	var removeModified bool
	removeErr := cmd.waitRemoveRepos(removeDone, removeCount, func(result *actionReposResult) {
		// Remove the repository from buildInfo
		buildInfo.Repos.removeByReposPath(result.repos.Path)
		removeModified = true
	})

	// Handle copy & remove errors
	if copyErr != nil || removeErr != nil {
		return multierror.Append(copyErr, removeErr).ErrorOrNil()
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

	// Write to build-info.json if buildInfo was modified
	if copyModified || removeModified {
		err = buildInfo.write()
		if err != nil {
			return err
		}
	}

	return nil
}

func (cmd *buildCmd) installRCFile(profileName, srcRCFileName, dst string, install bool) error {
	// Return error if destination file does not have magic comment
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
	src := filepath.Join(pathutil.RCDir(profileName), srcRCFileName)
	if !install || !pathutil.Exists(src) {
		return nil
	}

	return cmd.copyFileWithMagicComment(src, dst)
}

const magicComment = "\" NOTE: this file was generated by volt. please modify original file.\n"
const magicCommentNext = "\" Original file: %s\n\n"

// Return error if the magic comment does not exist
func (*buildCmd) shouldHaveMagicComment(dst string) error {
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

func (*buildCmd) copyFileWithMagicComment(src, dst string) (err error) {
	r, err := os.Open(src)
	if err != nil {
		return
	}
	defer func() {
		if e := r.Close(); e != nil {
			err = e
		}
	}()

	os.MkdirAll(filepath.Dir(dst), 0755)
	w, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := w.Close(); e != nil {
			err = e
		}
	}()

	_, err = w.Write([]byte(magicComment))
	if err != nil {
		return
	}
	_, err = w.Write([]byte(fmt.Sprintf(magicCommentNext, src)))
	if err != nil {
		return
	}

	_, err = io.Copy(w, r)
	return
}

type actionReposResult struct {
	err   error
	repos *lockjson.Repos
	files biFileMap
}

func (cmd *buildCmd) copyReposList(buildReposMap map[string]*biRepos, reposList []lockjson.Repos, optDir string) (chan actionReposResult, int) {
	copyDone := make(chan actionReposResult, len(reposList))
	copyCount := 0
	for i := range reposList {
		if reposList[i].Type == lockjson.ReposGitType {
			n, err := cmd.copyReposGit(&reposList[i], buildReposMap[reposList[i].Path], copyDone)
			if err != nil {
				copyDone <- actionReposResult{
					err:   errors.New("failed to copy " + string(reposList[i].Type) + " repos: " + err.Error()),
					repos: &reposList[i],
				}
			}
			copyCount += n
		} else if reposList[i].Type == lockjson.ReposStaticType {
			copyCount += cmd.copyReposStatic(&reposList[i], buildReposMap[reposList[i].Path], optDir, copyDone)
		} else {
			copyDone <- actionReposResult{
				err:   errors.New("invalid repository type: " + string(reposList[i].Type)),
				repos: &reposList[i],
			}
		}
	}
	return copyDone, copyCount
}

func (cmd *buildCmd) copyReposGit(repos *lockjson.Repos, buildRepos *biRepos, done chan actionReposResult) (int, error) {
	// Open ~/volt/repos/{repos}
	src := pathutil.FullReposPathOf(repos.Path)
	r, err := git.PlainOpen(src)
	if err != nil {
		return 0, errors.New("failed to open repository: " + err.Error())
	}

	cfg, err := r.Config()
	if err != nil {
		return 0, errors.New("failed to get repository config: " + err.Error())
	}

	isClean := false
	if wt, err := r.Worktree(); err == nil {
		if st, err := wt.Status(); err == nil && st.IsClean() {
			isClean = true
		}
	}

	if cmd.hasChangedGitRepos(repos, buildRepos, !isClean) {
		// Copy files from .git/objects/... when:
		// * bare repository
		// * or worktree is clean
		copyFromGitObjects := cfg.Core.IsBare || isClean
		go cmd.updateGitRepos(repos, r, copyFromGitObjects, done)
		return 1, nil
	}
	return 0, nil
}

func (cmd *buildCmd) copyReposStatic(repos *lockjson.Repos, buildRepos *biRepos, optDir string, done chan actionReposResult) int {
	if cmd.hasChangedStaticRepos(repos, buildRepos, optDir) {
		go cmd.updateStaticRepos(repos, done)
		return 1
	}
	return 0
}

// Remove vim repos found in lock.json, but not in build-info.json
func (cmd *buildCmd) removeReposList(buildInfoRepos biReposList, lockReposList lockjson.ReposList) (chan actionReposResult, int) {
	removeList := make(lockjson.ReposList, 0, len(lockReposList))
	for i := range lockReposList {
		if buildInfoRepos.findByReposPath(lockReposList[i].Path) == nil {
			removeList = append(removeList, lockReposList[i])
		}
	}
	removeDone := make(chan actionReposResult, len(removeList))
	for i := range removeList {
		go func(repos *lockjson.Repos) {
			// Remove directory under vim dir
			path := pathutil.PackReposPathOf(repos.Path)
			err := os.RemoveAll(path)
			logger.Info("Removing " + path + " ... Done.")
			removeDone <- actionReposResult{
				err:   err,
				repos: &lockjson.Repos{Path: repos.Path},
			}
		}(&removeList[i])
	}
	return removeDone, len(removeList)
}

func (*buildCmd) waitCopyRepos(copyDone chan actionReposResult, copyCount int, callback func(*actionReposResult) error) *multierror.Error {
	var merr *multierror.Error
	for i := 0; i < copyCount; i++ {
		result := <-copyDone
		if result.err != nil {
			merr = multierror.Append(
				merr,
				errors.New(
					"failed to copy repository '"+result.repos.Path+
						"': "+result.err.Error()))
		} else {
			err := callback(&result)
			if err != nil {
				merr = multierror.Append(merr, err)
			}
		}
	}
	return merr
}

func (*buildCmd) constructBuildInfo(buildInfo *buildInfoType, result *actionReposResult) {
	if result.repos.Type == lockjson.ReposGitType {
		r := buildInfo.Repos.findByReposPath(result.repos.Path)
		if r != nil {
			r.Version = result.repos.Version
			r.Files = result.files
		} else {
			buildInfo.Repos = append(
				buildInfo.Repos,
				biRepos{
					Type:    reposGitType,
					Path:    result.repos.Path,
					Version: result.repos.Version,
					Files:   result.files,
				},
			)
		}
	} else if result.repos.Type == lockjson.ReposStaticType {
		r := buildInfo.Repos.findByReposPath(result.repos.Path)
		if r != nil {
			r.Version = time.Now().Format(time.RFC3339)
			r.Files = result.files
		} else {
			buildInfo.Repos = append(
				buildInfo.Repos,
				biRepos{
					Type:    reposStaticType,
					Path:    result.repos.Path,
					Version: time.Now().Format(time.RFC3339),
					Files:   result.files,
				},
			)
		}
	} else {
		logger.Error("Unknown repos type (" + string(result.repos.Type) + ")")
	}
}

func (*buildCmd) waitRemoveRepos(removeDone chan actionReposResult, removeCount int, callback func(result *actionReposResult)) *multierror.Error {
	var merr *multierror.Error
	for i := 0; i < removeCount; i++ {
		result := <-removeDone
		if result.err != nil {
			merr = multierror.Append(
				merr, errors.New(
					"Failed to remove "+result.repos.Path+
						": "+result.err.Error()))
		} else {
			callback(&result)
		}
	}
	return merr
}

func (*buildCmd) getCurrentProfileAndReposList(lockJSON *lockjson.LockJSON) (*lockjson.Profile, []lockjson.Repos, error) {
	// Find current profile
	profile, err := lockJSON.Profiles.FindByName(lockJSON.CurrentProfileName)
	if err != nil {
		// this must not be occurred because lockjson.Read()
		// validates that the matching profile exists
		return nil, nil, err
	}

	reposList, err := lockJSON.GetReposListByProfile(profile)
	return profile, reposList, err
}

func (*buildCmd) getLatestModTime(path string) (time.Time, error) {
	mtime := time.Unix(0, 0)
	err := filepath.Walk(path, func(_ string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		t := fi.ModTime()
		if mtime.Before(t) {
			mtime = t
		}
		return nil
	})
	if err != nil {
		return time.Now(), errors.New("failed to readdir: " + err.Error())
	}
	return mtime, nil
}

func (*buildCmd) hasChangedGitRepos(repos *lockjson.Repos, buildRepos *biRepos, isDirty bool) bool {
	if buildRepos == nil { // Full build
		return true
	}
	if repos.Version != buildRepos.Version {
		return true
	}
	if buildRepos.DirtyWorktree || isDirty {
		return true
	}
	return false
}

// Remove ~/.vim/volt/opt/{repos} and copy from ~/volt/repos/{repos}
func (cmd *buildCmd) updateGitRepos(repos *lockjson.Repos, r *git.Repository, copyFromGitObjects bool, done chan actionReposResult) {
	src := pathutil.FullReposPathOf(repos.Path)
	dst := pathutil.PackReposPathOf(repos.Path)

	// Remove ~/.vim/volt/opt/{repos}
	// TODO: Do not remove here, copy newer files only after
	err := os.RemoveAll(dst)
	if err != nil {
		done <- actionReposResult{
			err:   errors.New("failed to remove repository: " + err.Error()),
			repos: repos,
		}
		return
	}

	if copyFromGitObjects {
		logger.Debug("Copy from git objects: " + repos.Path)
		cmd.updateBareGitRepos(r, src, dst, repos, done)
	} else {
		logger.Debug("Copy from filesystem: " + repos.Path)
		cmd.updateNonBareGitRepos(r, src, dst, repos, done)
	}
}

func (cmd *buildCmd) updateBareGitRepos(r *git.Repository, src, dst string, repos *lockjson.Repos, done chan actionReposResult) {
	// Get locked commit hash
	commit := plumbing.NewHash(repos.Version)
	commitObj, err := r.CommitObject(commit)
	if err != nil {
		done <- actionReposResult{
			err:   errors.New("failed to get HEAD commit object: " + err.Error()),
			repos: repos,
		}
		return
	}

	// Get tree hash of commit hash
	tree, err := r.TreeObject(commitObj.TreeHash)
	if err != nil {
		done <- actionReposResult{
			err:   errors.New("failed to get tree " + commit.String() + ": " + err.Error()),
			repos: repos,
		}
		return
	}

	// Copy files
	files := make(biFileMap, 512)
	err = tree.Files().ForEach(func(file *object.File) error {
		osMode, err := file.Mode.ToOSFileMode()
		if err != nil {
			return errors.New("failed to convert file mode: " + err.Error())
		}

		contents, err := file.Contents()
		if err != nil {
			return errors.New("failed to get file contents: " + err.Error())
		}

		filename := filepath.Join(dst, file.Name)
		os.MkdirAll(filepath.Dir(filename), 0755)
		ioutil.WriteFile(filename, []byte(contents), osMode)

		files[file.Name] = file.Hash.String() // blob hash
		return nil
	})
	if err != nil {
		done <- actionReposResult{
			err:   err,
			repos: repos,
		}
		return
	}

	// Do ":helptags" to generate tags file
	err = cmd.helptags(repos.Path)
	if err != nil {
		done <- actionReposResult{
			err:   err,
			repos: repos,
		}
		return
	}

	done <- actionReposResult{
		err:   nil,
		repos: repos,
		files: files,
	}
}

func (cmd *buildCmd) updateNonBareGitRepos(r *git.Repository, src, dst string, repos *lockjson.Repos, done chan actionReposResult) {
	files, err := ioutil.ReadDir(src)
	if err != nil {
		done <- actionReposResult{
			err:   err,
			repos: repos,
		}
		return
	}

	buf := make([]byte, 32*1024)
	created := make(map[string]bool, len(files))
	for _, file := range files {
		// Skip ".git" and ".gitignore"
		if file.Name() == ".git" || file.Name() == ".gitignore" {
			continue
		}
		if file.Mode()&BuildModeInvalidType != 0 {
			abspath := filepath.Join(src, file.Name())
			done <- actionReposResult{
				err:   errors.New(ErrBuildModeType + ": " + abspath),
				repos: repos,
			}
			return
		}
		if !created[dst] {
			os.MkdirAll(dst, 0755)
			created[dst] = true
		}
		from := filepath.Join(src, file.Name())
		to := filepath.Join(dst, file.Name())
		var err error
		if file.IsDir() {
			err = fileutil.TryLinkDir(from, to, buf, file.Mode(), BuildModeInvalidType)
		} else {
			err = fileutil.TryLinkFile(from, to, buf, file.Mode())
		}
		if err != nil {
			done <- actionReposResult{
				err:   err,
				repos: repos,
			}
			return
		}
	}

	err = cmd.helptags(repos.Path)
	if err != nil {
		done <- actionReposResult{
			err:   err,
			repos: repos,
		}
		return
	}

	done <- actionReposResult{
		err:   nil,
		repos: repos,
		files: nil, // all files are overwritten next time even when timestamp is older
	}
}

func (cmd *buildCmd) hasChangedStaticRepos(repos *lockjson.Repos, buildRepos *biRepos, optDir string) bool {
	if buildRepos == nil { // Full build
		return true
	}

	src := pathutil.FullReposPathOf(repos.Path)

	// Get latest mtime of src
	// TODO: Don't check mtime here, do it when copy altogether
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

// Remove ~/.vim/volt/opt/{repos} and copy from ~/volt/repos/{repos}
func (cmd *buildCmd) updateStaticRepos(repos *lockjson.Repos, done chan actionReposResult) {
	src := pathutil.FullReposPathOf(repos.Path)
	dst := pathutil.PackReposPathOf(repos.Path)

	// Remove ~/.vim/volt/opt/{repos}
	// TODO: Do not remove here, copy newer files only after
	err := os.RemoveAll(dst)
	if err != nil {
		done <- actionReposResult{
			err:   errors.New("failed to remove repository: " + err.Error()),
			repos: repos,
		}
		return
	}

	// Copy ~/volt/repos/{repos} to ~/.vim/volt/opt/{repos}
	buf := make([]byte, 32*1024)
	si, err := os.Stat(src)
	if err != nil {
		done <- actionReposResult{
			err:   errors.New("failed to copy static directory: " + err.Error()),
			repos: repos,
		}
		return
	}
	if !si.IsDir() {
		done <- actionReposResult{
			err:   errors.New("failed to copy static directory: source is not a directory"),
			repos: repos,
		}
		return
	}
	err = fileutil.TryLinkDir(src, dst, buf, si.Mode(), BuildModeInvalidType)
	if err != nil {
		done <- actionReposResult{
			err:   errors.New("failed to copy static directory: " + err.Error()),
			repos: repos,
		}
		return
	}

	// Do ":helptags" to generate tags file
	err = cmd.helptags(repos.Path)
	if err != nil {
		done <- actionReposResult{
			err:   err,
			repos: repos,
		}
		return
	}

	done <- actionReposResult{
		err:   nil,
		repos: repos,
	}
}

func (cmd *buildCmd) helptags(reposPath string) error {
	// Do nothing if <reposPath>/doc directory doesn't exist
	docdir := filepath.Join(pathutil.PackReposPathOf(reposPath), "doc")
	if !pathutil.Exists(docdir) {
		return nil
	}
	// Do not invoke vim if not installed
	_, err := exec.LookPath("vim")
	if err != nil {
		return errors.New("vim command is not in PATH: " + err.Error())
	}

	// Find vim executable from PATH
	vimExe, err := pathutil.VimExecutable()
	if err != nil {
		return err
	}
	vimArgs := cmd.makeVimArgs(reposPath)

	// Execute ":helptags doc" in reposPath
	logger.Debugf("Executing '%s %s' ...", vimExe, strings.Join(vimArgs, " "))
	err = exec.Command(vimExe, vimArgs...).Run()
	if err != nil {
		return errors.New("failed to make tags file: " + err.Error())
	}
	return nil
}

func (*buildCmd) makeVimArgs(reposPath string) []string {
	return []string{
		"-u", "NONE", "-N",
		"-c", "cd " + pathutil.PackReposPathOf(reposPath),
		"-c", "set rtp+=" + pathutil.PackReposPathOf(reposPath),
		"-c", "helptags doc", "-c", "quit",
	}
}
