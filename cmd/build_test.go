package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/haya14busa/go-vimlparser"
	"github.com/vim-volt/volt/fileutil"
	"github.com/vim-volt/volt/internal/testutil"
	"github.com/vim-volt/volt/pathutil"
)

// Checks:
// (A) Does not show `[ERROR]`, `[WARN]` messages
// (B) Exit with zero status
// (C) Do smart build
// (D) Do full build
// (E) `$VOLTPATH/repos/<repos>/` is copied to `~/.vim/pack/volt/<repos>/` (timestamp comparison)
// (F) vimrc with magic comment is installed
// (G) gvimrc with magic comment is installed
// (H) Installed bundled plugconf exists
// (I) Installed bundled plugconf is syntax OK

// * Run `volt build` (repos: exists, vim repos: not exist) (git repository) (A, B, C, E, H, I)
// * Put `$VOLTPATH/rc/<profile>/vimrc.vim` (F)
// * Put `$VOLTPATH/rc/<profile>/gvimrc.vim` (G)
func TestVoltBuildGitNoVimRepos(t *testing.T) {
	voltBuildGitNoVimRepos(t, false)
}

// * Run `volt build -full` (repos: exists, vim repos: not exist) (git repository) (A, B, D, E, H, I)
// * Put `$VOLTPATH/rc/<profile>/vimrc.vim` (F)
// * Put `$VOLTPATH/rc/<profile>/gvimrc.vim` (G)
func TestVoltBuildFullGitNoVimRepos(t *testing.T) {
	voltBuildGitNoVimRepos(t, true)
}

func voltBuildGitNoVimRepos(t *testing.T, full bool) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)
	reposPathList := []string{"github.com/tyru/caw.vim"}
	setUpTestdata(t, "caw.vim", reposGitType, reposPathList)
	rclist := installRCList(t, true, true, "default")

	// =============== run =============== //

	args := []string{"build"}
	if full {
		args = append(args, "-full")
	}
	out, err := testutil.RunVolt(args...)
	// (A, B)
	testutil.SuccessExit(t, out, err)

	// (C) and (D)
	checkBuildOutput(t, full, out)

	for _, reposPath := range reposPathList {
		// (E)
		checkCopied(t, reposPath)
	}

	// (F, G)
	checkRCInstalled(t, rclist)

	// (H)
	bundledPlugconf := pathutil.BundledPlugConf()
	if !pathutil.Exists(bundledPlugconf) {
		t.Fatalf("%s does not exist", bundledPlugconf)
	}

	// (I)
	checkSyntax(t, bundledPlugconf)
}

// * Run `volt build` (repos: newer, vim repos: older) (git repository) (A, B, C, E, H, I)
// * Put `$VOLTPATH/rc/<profile>/vimrc.vim` (F)
// * Put *no* `$VOLTPATH/rc/<profile>/gvimrc.vim` (!G)
func TestVoltBuildGitVimDirOlder(t *testing.T) {
	voltBuildGitVimDirOlder(t, false)
}

// * Run `volt build -full` (repos: newer, vim repos: older) (git repository) (A, B, D, E, H, I)
// * Put `$VOLTPATH/rc/<profile>/vimrc.vim` (F)
// * Put *no* `$VOLTPATH/rc/<profile>/gvimrc.vim` (!G)
func TestVoltBuildFullGitVimDirOlder(t *testing.T) {
	voltBuildGitVimDirOlder(t, true)
}

func voltBuildGitVimDirOlder(t *testing.T, full bool) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)
	reposPathList := []string{"github.com/tyru/caw.vim"}
	setUpTestdata(t, "caw.vim", reposGitType, reposPathList)
	out, err := testutil.RunVolt("build")
	testutil.SuccessExit(t, out, err)
	for _, reposPath := range reposPathList {
		touchFiles(t, pathutil.FullReposPathOf(reposPath))
	}
	rclist := installRCList(t, true, false, "default")

	// =============== run =============== //

	args := []string{"build"}
	if full {
		args = append(args, "-full")
	}
	out, err = testutil.RunVolt(args...)
	// (A, B)
	testutil.SuccessExit(t, out, err)

	// (C) and (D)
	checkBuildOutput(t, full, out)

	for _, reposPath := range reposPathList {
		// (E)
		checkCopied(t, reposPath)
	}

	// (F, !G)
	checkRCInstalled(t, rclist)

	// (H)
	bundledPlugconf := pathutil.BundledPlugConf()
	if !pathutil.Exists(bundledPlugconf) {
		t.Fatalf("%s does not exist", bundledPlugconf)
	}

	// (I)
	checkSyntax(t, bundledPlugconf)
}

// * Run `volt build` (repos: older, vim repos: newer) (git repository) (A, B, C, E, H, I)
// * Put *no* `$VOLTPATH/rc/<profile>/vimrc.vim` (!F)
// * Put `$VOLTPATH/rc/<profile>/gvimrc.vim` (G)
func TestVoltBuildGitVimDirNewer(t *testing.T) {
	voltBuildGitVimDirNewer(t, false)
}

// * Run `volt build -full` (repos: older, vim repos: newer) (git repository) (A, B, D, E, H, I)
// * Put *no* `$VOLTPATH/rc/<profile>/vimrc.vim` (!F)
// * Put `$VOLTPATH/rc/<profile>/gvimrc.vim` (G)
func TestVoltBuildFullGitVimDirNewer(t *testing.T) {
	voltBuildGitVimDirNewer(t, true)
}

func voltBuildGitVimDirNewer(t *testing.T, full bool) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)
	reposPathList := []string{"github.com/tyru/caw.vim"}
	setUpTestdata(t, "caw.vim", reposGitType, reposPathList)
	out, err := testutil.RunVolt("build")
	testutil.SuccessExit(t, out, err)
	for _, reposPath := range reposPathList {
		touchFiles(t, pathutil.PackReposPathOf(reposPath))
	}
	rclist := installRCList(t, false, true, "default")

	// =============== run =============== //

	args := []string{"build"}
	if full {
		args = append(args, "-full")
	}
	out, err = testutil.RunVolt(args...)
	// (A, B)
	testutil.SuccessExit(t, out, err)

	// (C) and (D)
	checkBuildOutput(t, full, out)

	for _, reposPath := range reposPathList {
		// (E)
		checkCopied(t, reposPath)
	}

	// (!F, G)
	checkRCInstalled(t, rclist)

	// (H)
	bundledPlugconf := pathutil.BundledPlugConf()
	if !pathutil.Exists(bundledPlugconf) {
		t.Fatalf("%s does not exist", bundledPlugconf)
	}

	// (I)
	checkSyntax(t, bundledPlugconf)
}

// * Run `volt build` (repos: exists, vim repos: not exist) (static repository) (A, B, C, E, H, I)
// * Put *no* `$VOLTPATH/rc/<profile>/vimrc.vim` (!F)
// * Put *no* `$VOLTPATH/rc/<profile>/gvimrc.vim` (!G)
func TestVoltBuildStaticNoVimRepos(t *testing.T) {
	voltBuildStaticNoVimRepos(t, false)
}

// * Run `volt build -full` (repos: exists, vim repos: not exist) (static repository) (A, B, D, E, H, I)
// * Put *no* `$VOLTPATH/rc/<profile>/vimrc.vim` (!F)
// * Put *no* `$VOLTPATH/rc/<profile>/gvimrc.vim` (!G)
func TestVoltBuildFullStaticNoVimRepos(t *testing.T) {
	voltBuildStaticNoVimRepos(t, true)
}

func voltBuildStaticNoVimRepos(t *testing.T, full bool) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)
	reposPathList := []string{"localhost/local/hello"}
	setUpTestdata(t, "hello", reposStaticType, reposPathList)
	rclist := installRCList(t, false, false, "default")

	// =============== run =============== //

	args := []string{"build"}
	if full {
		args = append(args, "-full")
	}
	out, err := testutil.RunVolt(args...)
	// (A, B)
	testutil.SuccessExit(t, out, err)

	// (C) and (D)
	checkBuildOutput(t, full, out)

	for _, reposPath := range reposPathList {
		// (E)
		checkCopied(t, reposPath)
	}

	// (!F, !G)
	checkRCInstalled(t, rclist)

	// (H)
	bundledPlugconf := pathutil.BundledPlugConf()
	if !pathutil.Exists(bundledPlugconf) {
		t.Fatalf("%s does not exist", bundledPlugconf)
	}

	// (I)
	checkSyntax(t, bundledPlugconf)
}

// * Run `volt build` (repos: newer, vim repos: older) (static repository) (A, B, C, E, H, I)
// * Put *no* `$VOLTPATH/rc/<profile>/vimrc.vim` (!F)
// * Put *no* `$VOLTPATH/rc/<profile>/gvimrc.vim` (!G)
func TestVoltBuildStaticVimDirOlder(t *testing.T) {
	voltBuildStaticVimDirOlder(t, false)
}

// * Run `volt build -full` (repos: newer, vim repos: older) (static repository) (A, B, D, E, H, I)
// * Put *no* `$VOLTPATH/rc/<profile>/vimrc.vim` (!F)
// * Put *no* `$VOLTPATH/rc/<profile>/gvimrc.vim` (!G)
func TestVoltBuildFullStaticVimDirOlder(t *testing.T) {
	voltBuildStaticVimDirOlder(t, true)
}

func voltBuildStaticVimDirOlder(t *testing.T, full bool) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)
	reposPathList := []string{"localhost/local/hello"}
	setUpTestdata(t, "hello", reposStaticType, reposPathList)
	out, err := testutil.RunVolt("build")
	testutil.SuccessExit(t, out, err)
	for _, reposPath := range reposPathList {
		touchFiles(t, pathutil.FullReposPathOf(reposPath))
	}
	rclist := installRCList(t, false, false, "default")

	// =============== run =============== //

	args := []string{"build"}
	if full {
		args = append(args, "-full")
	}
	out, err = testutil.RunVolt(args...)
	// (A, B)
	testutil.SuccessExit(t, out, err)

	// (C) and (D)
	checkBuildOutput(t, full, out)

	for _, reposPath := range reposPathList {
		// (E)
		checkCopied(t, reposPath)
	}

	// (!F, !G)
	checkRCInstalled(t, rclist)

	// (H)
	bundledPlugconf := pathutil.BundledPlugConf()
	if !pathutil.Exists(bundledPlugconf) {
		t.Fatalf("%s does not exist", bundledPlugconf)
	}

	// (I)
	checkSyntax(t, bundledPlugconf)
}

// * Run `volt build` (repos: older, vim repos: newer) (static repository) (A, B, C, E, H, I)
// * Put *no* `$VOLTPATH/rc/<profile>/vimrc.vim` (!F)
// * Put *no* `$VOLTPATH/rc/<profile>/gvimrc.vim` (!G)
func TestVoltBuildStaticVimDirNewer(t *testing.T) {
	voltBuildStaticVimDirNewer(t, false)
}

// * Run `volt build -full` (repos: older, vim repos: newer) (static repository) (A, B, D, E, H, I)
// * Put *no* `$VOLTPATH/rc/<profile>/vimrc.vim` (!F)
// * Put *no* `$VOLTPATH/rc/<profile>/gvimrc.vim` (!G)
func TestVoltBuildFullStaticVimDirNewer(t *testing.T) {
	voltBuildStaticVimDirNewer(t, true)
}

func voltBuildStaticVimDirNewer(t *testing.T, full bool) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)
	reposPathList := []string{"localhost/local/hello"}
	setUpTestdata(t, "hello", reposStaticType, reposPathList)
	out, err := testutil.RunVolt("build")
	testutil.SuccessExit(t, out, err)
	for _, reposPath := range reposPathList {
		touchFiles(t, pathutil.PackReposPathOf(reposPath))
	}
	rclist := installRCList(t, false, false, "default")

	// =============== run =============== //

	args := []string{"build"}
	if full {
		args = append(args, "-full")
	}
	out, err = testutil.RunVolt(args...)
	// (A, B)
	testutil.SuccessExit(t, out, err)

	// (C) and (D)
	checkBuildOutput(t, full, out)

	for _, reposPath := range reposPathList {
		// (E)
		checkCopied(t, reposPath)
	}

	// (!F, !G)
	checkRCInstalled(t, rclist)

	// (H)
	bundledPlugconf := pathutil.BundledPlugConf()
	if !pathutil.Exists(bundledPlugconf) {
		t.Fatalf("%s does not exist", bundledPlugconf)
	}

	// (I)
	checkSyntax(t, bundledPlugconf)
}

// `~/.vim/vimrc` with no magic comment exists (!A, !B, !E, !F, !G, !H)
func TestVoltBuildVimrcExists(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)
	reposPathList := []string{"github.com/tyru/caw.vim"}
	setUpTestdata(t, "caw.vim", reposGitType, reposPathList)
	rclist := installRCList(t, false, false, "default")

	vimrc := filepath.Join(pathutil.VimDir(), pathutil.Vimrc)
	os.MkdirAll(filepath.Dir(vimrc), 0777)
	if err := ioutil.WriteFile(vimrc, []byte("syntax on"), 0777); err != nil {
		t.Fatalf("cannot create %s: %s", vimrc, err.Error())
	}

	// =============== run =============== //

	out, err := testutil.RunVolt("build")
	// (!A, !B)
	testutil.FailExit(t, out, err)

	for _, reposPath := range reposPathList {
		// (!E)
		vimReposDir := pathutil.PackReposPathOf(reposPath)
		if pathutil.Exists(vimReposDir) {
			t.Fatalf("vim repos dir was created: %s", vimReposDir)
		}
	}

	// (!F, !G)
	checkRCInstalled(t, rclist)

	// (!H)
	bundledPlugconf := pathutil.BundledPlugConf()
	if pathutil.Exists(bundledPlugconf) {
		t.Fatalf("%s exists", bundledPlugconf)
	}
}

// `~/.vim/gvimrc` with no magic comment exists (!A, !B, !E, !F, !G, !H)
func TestVoltBuildGvimrcExists(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)
	reposPathList := []string{"github.com/tyru/caw.vim"}
	setUpTestdata(t, "caw.vim", reposGitType, reposPathList)
	rclist := installRCList(t, false, false, "default")

	gvimrc := filepath.Join(pathutil.VimDir(), pathutil.Gvimrc)
	os.MkdirAll(filepath.Dir(gvimrc), 0777)
	if err := ioutil.WriteFile(gvimrc, []byte("syntax on"), 0777); err != nil {
		t.Fatalf("cannot create %s: %s", gvimrc, err.Error())
	}

	// =============== run =============== //

	out, err := testutil.RunVolt("build")
	// (!A, !B)
	testutil.FailExit(t, out, err)

	for _, reposPath := range reposPathList {
		// (!E)
		vimReposDir := pathutil.PackReposPathOf(reposPath)
		if pathutil.Exists(vimReposDir) {
			t.Fatalf("vim repos dir was created: %s", vimReposDir)
		}
	}

	// (!F, !G)
	checkRCInstalled(t, rclist)

	// (!H)
	bundledPlugconf := pathutil.BundledPlugConf()
	if pathutil.Exists(bundledPlugconf) {
		t.Fatalf("%s exists", bundledPlugconf)
	}
}

// `~/.vim/vimrc` and `~/.vim/gvimrc` with no magic comment exists (!A, !B, !E, !F, !G, !H)
func TestVoltBuildVimrcAndGvimrcExists(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)
	reposPathList := []string{"github.com/tyru/caw.vim"}
	setUpTestdata(t, "caw.vim", reposGitType, reposPathList)
	rclist := installRCList(t, false, false, "default")

	vimrc := filepath.Join(pathutil.VimDir(), pathutil.Vimrc)
	os.MkdirAll(filepath.Dir(vimrc), 0777)
	if err := ioutil.WriteFile(vimrc, []byte("syntax on"), 0777); err != nil {
		t.Fatalf("cannot create %s: %s", vimrc, err.Error())
	}
	gvimrc := filepath.Join(pathutil.VimDir(), pathutil.Gvimrc)
	os.MkdirAll(filepath.Dir(gvimrc), 0777)
	if err := ioutil.WriteFile(gvimrc, []byte("syntax on"), 0777); err != nil {
		t.Fatalf("cannot create %s: %s", gvimrc, err.Error())
	}

	// =============== run =============== //

	out, err := testutil.RunVolt("build")
	// (!A, !B)
	testutil.FailExit(t, out, err)

	for _, reposPath := range reposPathList {
		// (!E)
		vimReposDir := pathutil.PackReposPathOf(reposPath)
		if pathutil.Exists(vimReposDir) {
			t.Fatalf("vim repos dir was created: %s", vimReposDir)
		}
	}

	// (!F, !G)
	checkRCInstalled(t, rclist)

	// (!H)
	bundledPlugconf := pathutil.BundledPlugConf()
	if pathutil.Exists(bundledPlugconf) {
		t.Fatalf("%s exists", bundledPlugconf)
	}
}

// ============================================

var testdataDir string

func init() {
	const thisFile = "cmd/build_test.go"
	_, fn, _, _ := runtime.Caller(0)
	dir := strings.TrimSuffix(fn, thisFile)
	testdataDir = filepath.Join(dir, "testdata")

	os.RemoveAll(filepath.Join(testdataDir, "voltpath"))
}

// Set up $VOLTPATH after "volt get <repos>"
// but the same repository is cloned only at first time
// under testdata/voltpath/{testdataName}/repos/<repos>
func setUpTestdata(t *testing.T, testdataName string, rType reposType, reposPathList []string) {
	voltpath := os.Getenv("VOLTPATH")
	tmpVoltpath := filepath.Join(testdataDir, "voltpath", testdataName)
	localSrcDir := filepath.Join(testdataDir, "local", testdataName)
	localName := fmt.Sprintf("localhost/local/%s", testdataName)
	buf := make([]byte, 32*1024)

	for _, reposPath := range reposPathList {
		testRepos := filepath.Join(tmpVoltpath, "repos", reposPath)
		if !pathutil.Exists(testRepos) {
			switch rType {
			case reposGitType:
				err := os.Setenv("VOLTPATH", tmpVoltpath)
				if err != nil {
					t.Fatal("failed to set VOLTPATH")
				}
				defer os.Setenv("VOLTPATH", voltpath)
				out, err := testutil.RunVolt("get", reposPath)
				testutil.SuccessExit(t, out, err)
			case reposStaticType:
				err := os.Setenv("VOLTPATH", tmpVoltpath)
				if err != nil {
					t.Fatal("failed to set VOLTPATH")
				}
				defer os.Setenv("VOLTPATH", voltpath)
				os.MkdirAll(filepath.Dir(testRepos), 0777)
				if err := fileutil.CopyDir(localSrcDir, testRepos, buf, 0777, 0); err != nil {
					t.Fatalf("failed to copy %s to %s", localSrcDir, testRepos)
				}
				out, err := testutil.RunVolt("get", localName)
				testutil.SuccessExit(t, out, err)
			default:
				t.Fatalf("unknown type %q", rType)
			}
		}

		// Copy repository
		repos := filepath.Join(voltpath, "repos", reposPath)
		os.MkdirAll(filepath.Dir(repos), 0777)
		if err := fileutil.CopyDir(testRepos, repos, buf, 0777, os.FileMode(0)); err != nil {
			t.Fatalf("failed to copy %s to %s", testRepos, repos)
		}

		// Copy lock.json
		testLockjsonPath := filepath.Join(tmpVoltpath, "lock.json")
		lockjsonPath := filepath.Join(voltpath, "lock.json")
		os.MkdirAll(filepath.Dir(lockjsonPath), 0777)
		if err := fileutil.CopyFile(testLockjsonPath, lockjsonPath, buf, 0777); err != nil {
			t.Fatalf("failed to copy %s to %s", testLockjsonPath, lockjsonPath)
		}
	}
}

func installRC(t *testing.T, file, profileName string) {
	src := filepath.Join(testdataDir, "rc", file)
	dst := filepath.Join(pathutil.RCDir(profileName), file)
	os.MkdirAll(filepath.Dir(dst), 0777)
	if err := fileutil.CopyFile(src, dst, nil, 0777); err != nil {
		t.Fatalf("cannot copy %s to %s: %s", src, dst, err.Error())
	}
}

func touchFiles(t *testing.T, fullpath string) {
	filepath.Walk(fullpath, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		var mtime time.Time
		if st, err := os.Lstat(path); err != nil {
			t.Fatalf("os.Lstat(%q) failed: %s", path, err.Error())
		} else {
			mtime = st.ModTime()
		}
		atime := mtime
		if err = os.Chtimes(path, atime, mtime); err != nil {
			t.Fatalf("failed to change timestamp %q: %s", path, err.Error())
		}
		return nil
	})
}

func checkBuildOutput(t *testing.T, full bool, out []byte) {
	outstr := string(out)
	contains := strings.Contains(outstr, "Full building")
	if !full && contains {
		t.Fatal("expected smart build but done by full build: " + outstr)
	} else if full && !contains {
		t.Fatal("expected full build but done by smart build: " + outstr)
	}
}

func checkCopied(t *testing.T, reposPath string) {
	vimReposDir := pathutil.PackReposPathOf(reposPath)
	reposDir := pathutil.FullReposPathOf(reposPath)
	tagsFile := filepath.Join("doc", "tags")
	filepath.Walk(vimReposDir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}

		// symlinks should not be copied
		if fi.Mode()&os.ModeSymlink != 0 {
			t.Fatal("symlinks are copied: " + path)
		}
		rel, err := filepath.Rel(vimReposDir, path)
		if err != nil {
			t.Fatalf("failed to get relative path of %s: %s", rel, err.Error())
		}
		// doc/tags is created after copy
		if rel == tagsFile {
			return nil
		}
		// .git, .gitignore should not be copied
		if rel == ".git" || rel == ".gitignore" {
			t.Fatal(".git or .gitignore are copied: " + rel)
		}

		reposFile := filepath.Join(reposDir, rel)
		if !sameFile(t, path, reposFile) {
			t.Fatalf("%s and %s are not same", rel, reposFile)
		}
		return nil
	})
}

func sameFile(t *testing.T, f1, f2 string) bool {
	fi1, err := os.Lstat(f1)
	if err != nil {
		t.Fatalf("os.Lstat(%q) returned error: %s", f1, err.Error())
	}
	fi2, err := os.Lstat(f2)
	if err != nil {
		t.Fatalf("os.Lstat(%q) returned error: %s", f2, err.Error())
	}
	// Compare metadata
	if os.SameFile(fi1, fi2) {
		return true
	}
	// Compare content
	b1, err := ioutil.ReadFile(f1)
	if err != nil {
		t.Fatalf("cannot read %s: %s", f1, err.Error())
	}
	b2, err := ioutil.ReadFile(f2)
	if err != nil {
		t.Fatalf("cannot read %s: %s", f2, err.Error())
	}
	return bytes.Equal(b1, b2)
}

type rcList struct {
	nameInRC  string
	nameInVim string
	installed bool
}

func installRCList(t *testing.T, vimrc, gvimrc bool, profileName string) []rcList {
	rclist := []rcList{
		rcList{pathutil.ProfileVimrc, pathutil.Vimrc, vimrc},
		rcList{pathutil.ProfileGvimrc, pathutil.Gvimrc, gvimrc},
	}
	for _, rc := range rclist {
		if rc.installed {
			installRC(t, rc.nameInRC, profileName)
		}
	}
	return rclist
}

func checkRCInstalled(t *testing.T, rclist []rcList) {
	for _, rc := range rclist {
		path := filepath.Join(pathutil.VimDir(), rc.nameInVim)
		if rc.installed {
			if !pathutil.Exists(path) {
				t.Fatalf("%s was not installed: %s", rc.nameInVim, path)
			}
			if err := (&buildCmd{}).shouldHaveMagicComment(path); err != nil {
				t.Fatalf("%s does not have magic comment: %s", rc.nameInVim, err.Error())
			}
		}
		if !rc.installed && pathutil.Exists(path) &&
			(&buildCmd{}).shouldHaveMagicComment(path) == nil {
			t.Fatalf("%s was installed: %s", rc.nameInVim, path)
		}
	}
}

func checkSyntax(t *testing.T, bundledPlugconf string) {
	r, err := os.Open(bundledPlugconf)
	if err != nil {
		t.Fatalf("failed to open %s: %s", bundledPlugconf, err.Error())
	}
	_, err = vimlparser.ParseFile(r, bundledPlugconf, nil)
	if err != nil {
		t.Fatalf("failed to parse %s: %s", bundledPlugconf, err.Error())
	}
}
