package cmd

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/haya14busa/go-vimlparser"
	"github.com/vim-volt/volt/fileutil"
	"github.com/vim-volt/volt/internal/testutil"
	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/pathutil"
)

// Checks:
// (A) Does not show `[ERROR]`, `[WARN]` messages
// (B) Exit with zero status
// (C) Do smart build
// (D) Do full build
// (E) `$VOLTPATH/repos/<repos>/` is copied to `~/.vim/pack/volt/<repos>/` (timestamp comparison)
// (F) `~/.vim/vimrc` exists
// (G) `~/.vim/vimrc` has magic comment
// (H) `~/.vim/gvimrc` exists
// (I) `~/.vim/gvimrc` has magic comment
// (J) Installed bundled plugconf exists
// (K) Installed bundled plugconf is syntax OK

// About vimrc and gvimrc test cases (F, G, H, I)
//
// Pre-conditions:
// (a) profile vimrc exists | profile gvimrc exists
// (b) user vimrc exists | user gvimrc exists
// (c) user vimrc has *no* magic comment | user gvimrc has *no* magic comment
//     (user vimrc or gvimrc is not installed by volt)
// (c') user vimrc has magic comment | user gvimrc has magic comment
//     (user vimrc or gvimrc is installed by volt)
//
// (case t1) a & !b (expects F,G if profile vimrc exists, expects H,I if profile gvimrc exists)
//   * if profile vimrc/gvimrc exists, it's installed to `~/.vim/{vimrc,gvimrc}`
//   * if profile vimrc/gvimrc does not exist, it's removed from `~/.vim/{vimrc,gvimrc}`
//   * (the case for the users of profile feature)
// (case t2) b & c (expects F,!G if user vimrc has *no* magic comment, expects H,!I if user gvimrc has *no* magic comment)
//   * if a & both profile & user vimrc exist: error
//   * if a & both profile & user gvimrc exist: error
//   * install profile vimrc/gvimrc if user vimrc/gvimrc does not exist
//   * user vimrc/gvimrc are not changed if vimrc/gvimrc exists
//   * (the case for the non-users of profile feature)
// (case t3) b & c' (expects !F,!H)
//   * if a: user vimrc/gvimrc are installed to `~/.vim/{vimrc,gvimrc}`
//   * if !a: user vimrc/gvimrc are removed
//   * (the case for the users of profile feature)
// (case t4) !a & !b (expects !F,!H)
//   * no vimrc/gvimrc are installed to `~/.vim/{vimrc,gvimrc}`

// * (case t1) profile vimrc:exists
//             profile gvimrc:exists
//             user vimrc:not exist
//             user gvimrc:not exist
//             vimrc magic comment:N/A
//             gvimrc magic comment:N/A (F, G, H, I)
func TestVoltBuildT1ProfileVimrcGvimrcExists(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)

	installProfileRC(t, "default", "vimrc-nomagic.vim", pathutil.ProfileVimrc)
	installProfileRC(t, "default", "gvimrc-nomagic.vim", pathutil.ProfileGvimrc)

	// =============== run =============== //

	out, err := testutil.RunVolt("build")
	// (A, B)
	testutil.SuccessExit(t, out, err)

	// (F, G, H, I)
	checkRCInstalled(t, 1, 1, 1, 1)
}

// * (case t1) profile vimrc:exists
//             profile gvimrc:not exist
//             user vimrc:not exist
//             user gvimrc:not exist
//             vimrc magic comment:N/A
//             gvimrc magic comment:N/A (F, G, !H)
func TestVoltBuildT1ProfileVimrcExists(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)

	installProfileRC(t, "default", "vimrc-nomagic.vim", pathutil.ProfileVimrc)

	// =============== run =============== //

	out, err := testutil.RunVolt("build")
	// (A, B)
	testutil.SuccessExit(t, out, err)

	// (F, G, !H)
	checkRCInstalled(t, 1, 1, 0, -1)
}

// * (case t1) profile vimrc:not exist
//             profile gvimrc:exists
//             user vimrc:not exist
//             user gvimrc:not exist
//             vimrc magic comment:N/A
//             gvimrc magic comment:N/A (!F, H, I)
func TestVoltBuildT1ProfileGvimrcExists(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)

	installProfileRC(t, "default", "gvimrc-nomagic.vim", pathutil.ProfileGvimrc)

	// =============== run =============== //

	out, err := testutil.RunVolt("build")
	// (A, B)
	testutil.SuccessExit(t, out, err)

	// (!F, H, I)
	checkRCInstalled(t, 0, -1, 1, 1)
}

// * (case t2) profile vimrc:not exist
//             profile gvimrc:not exist
//             user vimrc:exists
//             user gvimrc:exists
//             vimrc magic comment:not exist
//             gvimrc magic comment:not exist (F, !G, H, !I)
func TestVoltBuildT2UserVimrcGvimrcExists(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)

	installVimRC(t, "vimrc-nomagic.vim", pathutil.Vimrc)
	installVimRC(t, "gvimrc-nomagic.vim", pathutil.Gvimrc)

	// =============== run =============== //

	out, err := testutil.RunVolt("build")
	// (A, B)
	testutil.SuccessExit(t, out, err)

	// (F, !G, H, !I)
	checkRCInstalled(t, 1, 0, 1, 0)
}

// * (case t2) profile vimrc:not exist
//             profile gvimrc:not exist
//             user vimrc:exists
//             user gvimrc:not exist
//             vimrc magic comment:not exist
//             gvimrc magic comment:N/A (F, !G, !H)
func TestVoltBuildT2UserVimrcExists(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)

	installVimRC(t, "vimrc-nomagic.vim", pathutil.Vimrc)

	// =============== run =============== //

	out, err := testutil.RunVolt("build")
	// (A, B)
	testutil.SuccessExit(t, out, err)

	// (F, !G, !H)
	checkRCInstalled(t, 1, 0, 0, -1)
}

// * Run `volt build` (!A, !B)
// * (case t2) profile vimrc:exists
//             profile gvimrc:not exist
//             user vimrc:exists
//             user gvimrc:not exist
//             vimrc magic comment:not exist
//             gvimrc magic comment:N/A (F, !G, !H)
func TestErrVoltBuildT2CannotOverwriteUserVimrc(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)

	installProfileRC(t, "default", "vimrc-nomagic.vim", pathutil.ProfileVimrc)
	installVimRC(t, "vimrc-nomagic.vim", pathutil.Vimrc)

	// =============== run =============== //

	out, err := testutil.RunVolt("build")
	// (!A, !B)
	testutil.FailExit(t, out, err)

	// (F, !G, !H)
	checkRCInstalled(t, 1, 0, 0, -1)
}

// * Run `volt build` (!A, !B)
// * (case t2) profile vimrc:not exist
//             profile gvimrc:exists
//             user vimrc:not exist
//             user gvimrc:exists
//             vimrc magic comment:N/A
//             gvimrc magic comment:not exist (!F, H, !I)
func TestErrVoltBuildT2CannotOverwriteUserGvimrc(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)

	installProfileRC(t, "default", "gvimrc-nomagic.vim", pathutil.ProfileGvimrc)
	installVimRC(t, "gvimrc-nomagic.vim", pathutil.Gvimrc)

	// =============== run =============== //

	out, err := testutil.RunVolt("build")
	// (!A, !B)
	testutil.FailExit(t, out, err)

	// (!F, H, !I)
	checkRCInstalled(t, 0, -1, 1, 0)
}

// * Run `volt build` (!A, !B)
// * (case t2) profile vimrc:exists
//             profile gvimrc:exists
//             user vimrc:not exist
//             user gvimrc:exists
//             vimrc magic comment:N/A
//             gvimrc magic comment:not exist (!F, H, !I)
func TestErrVoltBuildT2DontInstallVimrc(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)

	installProfileRC(t, "default", "vimrc-nomagic.vim", pathutil.ProfileVimrc)
	installProfileRC(t, "default", "gvimrc-nomagic.vim", pathutil.ProfileGvimrc)
	installVimRC(t, "gvimrc-nomagic.vim", pathutil.Gvimrc)

	// =============== run =============== //

	out, err := testutil.RunVolt("build")
	// (!A, !B)
	testutil.FailExit(t, out, err)

	// (!F, H, !I)
	checkRCInstalled(t, 0, -1, 1, 0)
}

// * Run `volt build` (!A, !B)
// * (case t2) profile vimrc:exists
//             profile gvimrc:exists
//             user vimrc:exists
//             user gvimrc:not exist
//             vimrc magic comment:not exist
//             gvimrc magic comment:N/A (F, !G, !H)
func TestErrVoltBuildT2DontInstallGvimrc(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)

	installProfileRC(t, "default", "vimrc-nomagic.vim", pathutil.ProfileVimrc)
	installProfileRC(t, "default", "gvimrc-nomagic.vim", pathutil.ProfileGvimrc)
	installVimRC(t, "vimrc-nomagic.vim", pathutil.Vimrc)

	// =============== run =============== //

	out, err := testutil.RunVolt("build")
	// (!A, !B)
	testutil.FailExit(t, out, err)

	// (F, !G, !H)
	checkRCInstalled(t, 1, 0, 0, -1)
}

// * Run `volt build` (A, B)
// * (case t2) profile vimrc:exists
//             profile gvimrc:not exist
//             user vimrc:not exist
//             user gvimrc:exists
//             vimrc magic comment:not exist
//             gvimrc magic comment:N/A (F, G, H, !I)
func TestVoltBuildT2CanInstallUserVimrc(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)

	installProfileRC(t, "default", "vimrc-nomagic.vim", pathutil.ProfileVimrc)
	installVimRC(t, "gvimrc-nomagic.vim", pathutil.Gvimrc)

	// =============== run =============== //

	out, err := testutil.RunVolt("build")
	// (A, B)
	testutil.SuccessExit(t, out, err)

	// (F, G, H, !I)
	checkRCInstalled(t, 1, 1, 1, 0)
}

// * Run `volt build` (A, B)
// * (case t3) profile vimrc:exists
//             profile gvimrc:exists
//             user vimrc:exists
//             user gvimrc:exists
//             vimrc magic comment:exists
//             gvimrc magic comment:exists (F, G, H, I)
func TestVoltBuildT3OverwriteUserVimrcGvimrcByProfileVimrcGvimrc(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)

	installProfileRC(t, "default", "vimrc-nomagic.vim", pathutil.ProfileVimrc)
	installProfileRC(t, "default", "gvimrc-nomagic.vim", pathutil.ProfileGvimrc)
	installVimRC(t, "vimrc-magic.vim", pathutil.Vimrc)
	installVimRC(t, "gvimrc-magic.vim", pathutil.Gvimrc)

	// =============== run =============== //

	out, err := testutil.RunVolt("build")
	// (A, B)
	testutil.SuccessExit(t, out, err)

	// (F, G, H, I)
	checkRCInstalled(t, 1, 1, 1, 1)
}

// * Run `volt build` (A, B)
// * (case t3) profile vimrc:not exist
//             profile gvimrc:exists
//             user vimrc:not exist
//             user gvimrc:exists
//             vimrc magic comment:N/A
//             gvimrc magic comment:exists (!F, H, I)
func TestVoltBuildT3OverwriteUserGvimrcByProfileGvimrc(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)

	installProfileRC(t, "default", "gvimrc-nomagic.vim", pathutil.ProfileGvimrc)
	installVimRC(t, "gvimrc-magic.vim", pathutil.Gvimrc)

	// =============== run =============== //

	out, err := testutil.RunVolt("build")
	// (A, B)
	testutil.SuccessExit(t, out, err)

	// (!F, H, I)
	checkRCInstalled(t, 0, -1, 1, 1)
}

// * Run `volt build` (A, B)
// * (case t3) profile vimrc:exists
//             profile gvimrc:not exist
//             user vimrc:exists
//             user gvimrc:not exist
//             vimrc magic comment:exists
//             gvimrc magic comment:N/A (F, G, !H)
func TestVoltBuildT3OverwriteUserVimrcByProfileVimrc(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)

	installProfileRC(t, "default", "vimrc-nomagic.vim", pathutil.ProfileVimrc)
	installVimRC(t, "vimrc-magic.vim", pathutil.Vimrc)

	// =============== run =============== //

	out, err := testutil.RunVolt("build")
	// (A, B)
	testutil.SuccessExit(t, out, err)

	// (F, G, !H)
	checkRCInstalled(t, 1, 1, 0, -1)
}

// * Run `volt build` (A, B)
// * (case t3) profile vimrc:not exist
//             profile gvimrc:not exist
//             user vimrc:exists
//             user gvimrc:exists
//             vimrc magic comment:exists
//             gvimrc magic comment:exists (!F, !H)
func TestVoltBuildT3RemoveUserVimrcGvimrc(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)

	installVimRC(t, "vimrc-magic.vim", pathutil.Vimrc)
	installVimRC(t, "gvimrc-magic.vim", pathutil.Gvimrc)

	// =============== run =============== //

	out, err := testutil.RunVolt("build")
	// (A, B)
	testutil.SuccessExit(t, out, err)

	// (!F, !H)
	checkRCInstalled(t, 0, -1, 0, -1)
}

// * Run `volt build` (A, B)
// * (case t3) profile vimrc:not exist
//             profile gvimrc:exists
//             user vimrc:exists
//             user gvimrc:not exist
//             vimrc magic comment:exists
//             gvimrc magic comment:N/A (!F, H, I)
func TestVoltBuildT3InstallGvimrcAndRemoveUserVimrc(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)

	installProfileRC(t, "default", "gvimrc-nomagic.vim", pathutil.ProfileGvimrc)
	installVimRC(t, "vimrc-magic.vim", pathutil.Vimrc)

	// =============== run =============== //

	out, err := testutil.RunVolt("build")
	// (A, B)
	testutil.SuccessExit(t, out, err)

	// (!F, H, I)
	checkRCInstalled(t, 0, -1, 1, 1)
}

// * Run `volt build` (A, B)
// * (case t3) profile vimrc:exists
//             profile gvimrc:not exist
//             user vimrc:not exist
//             user gvimrc:exists
//             vimrc magic comment:N/A
//             gvimrc magic comment:exists (F, G, !H)
func TestVoltBuildT3InstallVimrcAndRemoveUserGvimrc(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)

	installProfileRC(t, "default", "vimrc-nomagic.vim", pathutil.ProfileVimrc)
	installVimRC(t, "gvimrc-magic.vim", pathutil.Gvimrc)

	// =============== run =============== //

	out, err := testutil.RunVolt("build")
	// (A, B)
	testutil.SuccessExit(t, out, err)

	// (F, G, !H)
	checkRCInstalled(t, 1, 1, 0, -1)
}

// * Run `volt build` (A, B)
// * (case t4) profile vimrc:not exist
//             profile gvimrc:not exist
//             user vimrc:not exist
//             user gvimrc:not exist
//             vimrc magic comment:N/A
//             gvimrc magic comment:N/A (!F, !H)
func TestVoltBuildT4NoVimrcGvimrc(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)

	// =============== run =============== //

	out, err := testutil.RunVolt("build")
	// (A, B)
	testutil.SuccessExit(t, out, err)

	// (!F, !H)
	checkRCInstalled(t, 0, -1, 0, -1)
}

// ===========================================================

// * Run `volt build` (repos: exists, vim repos: not exist) (git repository) (A, B, C, E, !F, !H, J, K)
func TestVoltBuildGitNoVimRepos(t *testing.T) {
	voltBuildGitNoVimRepos(t, false)
}

// * Run `volt build -full` (repos: exists, vim repos: not exist) (git repository) (A, B, D, E, !F, !H, J, K)
func TestVoltBuildFullGitNoVimRepos(t *testing.T) {
	voltBuildGitNoVimRepos(t, true)
}

func voltBuildGitNoVimRepos(t *testing.T, full bool) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)
	reposPathList := []string{"github.com/tyru/caw.vim"}
	testutil.SetUpRepos(t, "caw.vim", lockjson.ReposGitType, reposPathList)

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

	// (!F, !H)
	checkRCInstalled(t, 0, -1, 0, -1)

	// (J)
	bundledPlugconf := pathutil.BundledPlugConf()
	if !pathutil.Exists(bundledPlugconf) {
		t.Fatalf("%s does not exist", bundledPlugconf)
	}

	// (K)
	checkSyntax(t, bundledPlugconf)
}

// * Run `volt build` (repos: newer, vim repos: older) (git repository) (A, B, C, E, !F, !H, J, K)
func TestVoltBuildGitVimDirOlder(t *testing.T) {
	voltBuildGitVimDirOlder(t, false)
}

// * Run `volt build -full` (repos: newer, vim repos: older) (git repository) (A, B, D, E, !F, !H, J, K)
func TestVoltBuildFullGitVimDirOlder(t *testing.T) {
	voltBuildGitVimDirOlder(t, true)
}

func voltBuildGitVimDirOlder(t *testing.T, full bool) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)
	reposPathList := []string{"github.com/tyru/caw.vim"}
	testutil.SetUpRepos(t, "caw.vim", lockjson.ReposGitType, reposPathList)
	out, err := testutil.RunVolt("build")
	testutil.SuccessExit(t, out, err)
	for _, reposPath := range reposPathList {
		touchFiles(t, pathutil.FullReposPathOf(reposPath))
	}

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

	// (!F, !H)
	checkRCInstalled(t, 0, -1, 0, -1)

	// (J)
	bundledPlugconf := pathutil.BundledPlugConf()
	if !pathutil.Exists(bundledPlugconf) {
		t.Fatalf("%s does not exist", bundledPlugconf)
	}

	// (K)
	checkSyntax(t, bundledPlugconf)
}

// * Run `volt build` (repos: older, vim repos: newer) (git repository) (A, B, C, E, !F, !H, J, K)
func TestVoltBuildGitVimDirNewer(t *testing.T) {
	voltBuildGitVimDirNewer(t, false)
}

// * Run `volt build -full` (repos: older, vim repos: newer) (git repository) (A, B, D, E, !F, !H, J, K)
func TestVoltBuildFullGitVimDirNewer(t *testing.T) {
	voltBuildGitVimDirNewer(t, true)
}

func voltBuildGitVimDirNewer(t *testing.T, full bool) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)
	reposPathList := []string{"github.com/tyru/caw.vim"}
	testutil.SetUpRepos(t, "caw.vim", lockjson.ReposGitType, reposPathList)
	out, err := testutil.RunVolt("build")
	testutil.SuccessExit(t, out, err)
	for _, reposPath := range reposPathList {
		touchFiles(t, pathutil.PackReposPathOf(reposPath))
	}

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

	// (!F, !H)
	checkRCInstalled(t, 0, -1, 0, -1)

	// (J)
	bundledPlugconf := pathutil.BundledPlugConf()
	if !pathutil.Exists(bundledPlugconf) {
		t.Fatalf("%s does not exist", bundledPlugconf)
	}

	// (K)
	checkSyntax(t, bundledPlugconf)
}

// * Run `volt build` (repos: exists, vim repos: not exist) (static repository) (A, B, C, E, !F, !H, J, K)
func TestVoltBuildStaticNoVimRepos(t *testing.T) {
	voltBuildStaticNoVimRepos(t, false)
}

// * Run `volt build -full` (repos: exists, vim repos: not exist) (static repository) (A, B, D, E, !F, !H, J, K)
func TestVoltBuildFullStaticNoVimRepos(t *testing.T) {
	voltBuildStaticNoVimRepos(t, true)
}

func voltBuildStaticNoVimRepos(t *testing.T, full bool) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)
	reposPathList := []string{"localhost/local/hello"}
	testutil.SetUpRepos(t, "hello", lockjson.ReposStaticType, reposPathList)

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

	// (!F, !H)
	checkRCInstalled(t, 0, -1, 0, -1)

	// (J)
	bundledPlugconf := pathutil.BundledPlugConf()
	if !pathutil.Exists(bundledPlugconf) {
		t.Fatalf("%s does not exist", bundledPlugconf)
	}

	// (K)
	checkSyntax(t, bundledPlugconf)
}

// * Run `volt build` (repos: newer, vim repos: older) (static repository) (A, B, C, E, !F, !H, J, K)
func TestVoltBuildStaticVimDirOlder(t *testing.T) {
	voltBuildStaticVimDirOlder(t, false)
}

// * Run `volt build -full` (repos: newer, vim repos: older) (static repository) (A, B, D, E, !F, !H, J, K)
func TestVoltBuildFullStaticVimDirOlder(t *testing.T) {
	voltBuildStaticVimDirOlder(t, true)
}

func voltBuildStaticVimDirOlder(t *testing.T, full bool) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)
	reposPathList := []string{"localhost/local/hello"}
	testutil.SetUpRepos(t, "hello", lockjson.ReposStaticType, reposPathList)
	out, err := testutil.RunVolt("build")
	testutil.SuccessExit(t, out, err)
	for _, reposPath := range reposPathList {
		touchFiles(t, pathutil.FullReposPathOf(reposPath))
	}

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

	// (!F, !H)
	checkRCInstalled(t, 0, -1, 0, -1)

	// (J)
	bundledPlugconf := pathutil.BundledPlugConf()
	if !pathutil.Exists(bundledPlugconf) {
		t.Fatalf("%s does not exist", bundledPlugconf)
	}

	// (K)
	checkSyntax(t, bundledPlugconf)
}

// * Run `volt build` (repos: older, vim repos: newer) (static repository) (A, B, C, E, !F, !H, J, K)
func TestVoltBuildStaticVimDirNewer(t *testing.T) {
	voltBuildStaticVimDirNewer(t, false)
}

// * Run `volt build -full` (repos: older, vim repos: newer) (static repository) (A, B, D, E, !F, !H, J, K)
func TestVoltBuildFullStaticVimDirNewer(t *testing.T) {
	voltBuildStaticVimDirNewer(t, true)
}

func voltBuildStaticVimDirNewer(t *testing.T, full bool) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)
	reposPathList := []string{"localhost/local/hello"}
	testutil.SetUpRepos(t, "hello", lockjson.ReposStaticType, reposPathList)
	out, err := testutil.RunVolt("build")
	testutil.SuccessExit(t, out, err)
	for _, reposPath := range reposPathList {
		touchFiles(t, pathutil.PackReposPathOf(reposPath))
	}

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

	// (!F, !H)
	checkRCInstalled(t, 0, -1, 0, -1)

	// (J)
	bundledPlugconf := pathutil.BundledPlugConf()
	if !pathutil.Exists(bundledPlugconf) {
		t.Fatalf("%s does not exist", bundledPlugconf)
	}

	// (K)
	checkSyntax(t, bundledPlugconf)
}

// ============================================

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

func installProfileRC(t *testing.T, profileName, srcName, dstName string) {
	src := filepath.Join(testutil.TestdataDir(), "rc", srcName)
	dst := filepath.Join(pathutil.RCDir(profileName), dstName)
	os.MkdirAll(filepath.Dir(dst), 0777)
	if err := fileutil.CopyFile(src, dst, nil, 0777); err != nil {
		t.Fatalf("cannot copy %s to %s: %s", src, dst, err.Error())
	}
}

func installVimRC(t *testing.T, srcName, dstName string) {
	src := filepath.Join(testutil.TestdataDir(), "rc", srcName)
	dst := filepath.Join(pathutil.VimDir(), dstName)
	os.MkdirAll(filepath.Dir(dst), 0777)
	if err := fileutil.CopyFile(src, dst, nil, 0777); err != nil {
		t.Fatalf("cannot copy %s to %s: %s", src, dst, err.Error())
	}
}

func checkRCInstalled(t *testing.T, f, g, h, i int) {
	userVimrc := filepath.Join(pathutil.VimDir(), pathutil.Vimrc)
	userGvimrc := filepath.Join(pathutil.VimDir(), pathutil.Gvimrc)

	// (F, H)
	for _, tt := range []struct {
		value int
		path  string
	}{
		{f, userVimrc},
		{h, userGvimrc},
	} {
		if tt.value >= 0 {
			if tt.value == 1 && !pathutil.Exists(tt.path) {
				t.Fatalf("expected %s was installed but not installed", tt.path)
			}
			if tt.value == 0 && pathutil.Exists(tt.path) {
				t.Fatalf("expected %s was not installed but installed", tt.path)
			}
		}
	}

	// (G, I)
	for _, tt := range []struct {
		value int
		path  string
	}{
		{g, userVimrc},
		{i, userGvimrc},
	} {
		if tt.value >= 0 {
			if tt.value == 1 && !(&buildCmd{}).hasMagicComment(tt.path) {
				t.Fatalf("expected %s has magic comment but has no magic comment", tt.path)
			}
			if tt.value == 0 && (&buildCmd{}).hasMagicComment(tt.path) {
				t.Fatalf("expected %s was not installed but installed", tt.path)
			}
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
