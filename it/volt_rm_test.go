package it

import (
	"os"
	"strings"
	"testing"

	"github.com/vim-volt/volt/internal/testutils"
	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/pathutil"
)

// Checks:
// (A) Does not show `[ERROR]`, `[WARN]` messages
// (B) Exit with zero status
// (C) Directories of `$VOLTPATH/repos/<repos>/` are removed
// (D) Plugconf of `$VOLTPATH/plugconf/<repos>.vim` are removed
// (E) Repositories are removed from `~/.vim/pack/volt/<repos>/`
// (F) Specified entries in lock.json are removed

// TODO: Add test cases
// * [error] Run `volt rm <plugin>` when the plugin is depended by some plugins (!A, !B, !C, !D, !E, !F)

// Run `volt rm <plugin>` (repos: exists, plugconf: exists, vim repos: exists) (A, B, C, !D, E, F)
func TestVoltRmOnePlugin(t *testing.T) {
	testutils.SetUpEnv(t)
	out, err := testutils.RunVolt("get", "tyru/caw.vim")
	testutils.SuccessExit(t, out, err)
	out, err = testutils.RunVolt("rm", "tyru/caw.vim")
	// (A, B)
	testutils.SuccessExit(t, out, err)
	reposPath := "github.com/tyru/caw.vim"

	// (C)
	reposDir := pathutil.FullReposPathOf(reposPath)
	if pathutil.Exists(reposDir) {
		t.Fatal("repos was not removed: " + reposDir)
	}

	// (!D)
	plugconf := pathutil.PlugconfOf(reposPath)
	if !pathutil.Exists(plugconf) {
		t.Fatal("plugconf was removed: " + plugconf)
	}

	// (E)
	vimReposDir := pathutil.PackReposPathOf(reposPath)
	if pathutil.Exists(vimReposDir) {
		t.Fatal("vim repos was not removed: " + vimReposDir)
	}

	// (F)
	testReposPathWereRemoved(t, reposPath)
}

// Run `volt rm <plugin>` (repos: exists, plugconf: not exists, vim repos: exists) (A, B, C, E, F)
func TestVoltRmOnePluginNoPlugconf(t *testing.T) {
	testutils.SetUpEnv(t)
	out, err := testutils.RunVolt("get", "tyru/caw.vim")
	testutils.SuccessExit(t, out, err)
	reposPath := "github.com/tyru/caw.vim"
	if err := os.Remove(pathutil.PlugconfOf(reposPath)); err != nil {
		t.Fatal("failed to remove plugconf: " + err.Error())
	}
	out, err = testutils.RunVolt("rm", "tyru/caw.vim")
	// (A, B)
	testutils.SuccessExit(t, out, err)

	// (C)
	reposDir := pathutil.FullReposPathOf(reposPath)
	if pathutil.Exists(reposDir) {
		t.Fatal("repos was not removed: " + reposDir)
	}

	// (E)
	vimReposDir := pathutil.PackReposPathOf(reposPath)
	if pathutil.Exists(vimReposDir) {
		t.Fatal("vim repos was not removed: " + vimReposDir)
	}

	// (F)
	testReposPathWereRemoved(t, reposPath)
}

// Run `volt rm <plugin1> <plugin2>` (repos: exists, plugconf: exists, vim repos: exists) (A, B, C, !D, E, F)
func TestVoltRmTwoOrMorePluginNoPlugconf(t *testing.T) {
	testutils.SetUpEnv(t)
	out, err := testutils.RunVolt("get", "tyru/caw.vim", "tyru/capture.vim")
	testutils.SuccessExit(t, out, err)
	out, err = testutils.RunVolt("rm", "tyru/caw.vim", "tyru/capture.vim")
	// (A, B)
	testutils.SuccessExit(t, out, err)
	cawReposPath := "github.com/tyru/caw.vim"
	captureReposPath := "github.com/tyru/capture.vim"

	for _, reposPath := range []string{cawReposPath, captureReposPath} {
		// (C)
		reposDir := pathutil.FullReposPathOf(reposPath)
		if pathutil.Exists(reposDir) {
			t.Fatal("repos was not removed: " + reposDir)
		}

		// (!D)
		plugconf := pathutil.PlugconfOf(reposPath)
		if !pathutil.Exists(plugconf) {
			t.Fatal("plugconf was removed: " + plugconf)
		}

		// (E)
		vimReposDir := pathutil.PackReposPathOf(reposPath)
		if pathutil.Exists(vimReposDir) {
			t.Fatal("vim repos was not removed: " + vimReposDir)
		}

		// (F)
		testReposPathWereRemoved(t, reposPath)
	}
}

// [error] Run `volt rm <plugin>` (repos: not exists, plugconf: exists, vim repos: exists) (!A, !B, !D)
func TestErrVoltRmOnePluginNoRepos(t *testing.T) {
	testutils.SetUpEnv(t)
	out, err := testutils.RunVolt("get", "tyru/caw.vim")
	testutils.SuccessExit(t, out, err)
	reposPath := "github.com/tyru/caw.vim"
	if err := os.RemoveAll(pathutil.FullReposPathOf(reposPath)); err != nil {
		t.Fatal("failed to remove repos: " + err.Error())
	}
	out, err = testutils.RunVolt("rm", "tyru/caw.vim")
	// (!A, !B)
	testutils.FailExit(t, out, err)

	// (!D)
	plugconf := pathutil.PlugconfOf(reposPath)
	if !pathutil.Exists(plugconf) {
		t.Fatal("plugconf was removed: " + plugconf)
	}

	// TODO: Show error message that repos and vim repos are mismatch
}

// [error] Run `volt rm <plugin>` (repos: not exists, plugconf: not exists, vim repos: exists) (!A, !B)
func TestErrVoltRmOnePluginNoReposNoPlugconf(t *testing.T) {
	testutils.SetUpEnv(t)
	out, err := testutils.RunVolt("get", "tyru/caw.vim")
	testutils.SuccessExit(t, out, err)
	reposPath := "github.com/tyru/caw.vim"
	if err := os.RemoveAll(pathutil.FullReposPathOf(reposPath)); err != nil {
		t.Fatal("failed to remove repos: " + err.Error())
	}
	if err := os.Remove(pathutil.PlugconfOf(reposPath)); err != nil {
		t.Fatal("failed to remove plugconf: " + err.Error())
	}
	out, err = testutils.RunVolt("rm", "tyru/caw.vim")
	// (!A, !B)
	testutils.FailExit(t, out, err)

	// TODO: Show error message that repos and vim repos are mismatch
}

// Run `volt rm -p <plugin>` (repos: exists, plugconf: exists, vim repos: exists) (A, B, C, D, E, F)
func TestVoltRmPoptOnePlugin(t *testing.T) {
	testutils.SetUpEnv(t)
	out, err := testutils.RunVolt("get", "tyru/caw.vim")
	testutils.SuccessExit(t, out, err)
	out, err = testutils.RunVolt("rm", "-p", "tyru/caw.vim")
	// (A, B)
	testutils.SuccessExit(t, out, err)
	reposPath := "github.com/tyru/caw.vim"

	// (C)
	reposDir := pathutil.FullReposPathOf(reposPath)
	if pathutil.Exists(reposDir) {
		t.Fatal("repos was not removed: " + reposDir)
	}

	// (D)
	plugconf := pathutil.PlugconfOf(reposPath)
	if pathutil.Exists(plugconf) {
		t.Fatal("plugconf was not removed: " + plugconf)
	}

	// (E)
	vimReposDir := pathutil.PackReposPathOf(reposPath)
	if pathutil.Exists(vimReposDir) {
		t.Fatal("vim repos was not removed: " + vimReposDir)
	}

	// (F)
	testReposPathWereRemoved(t, reposPath)
}

// Run `volt rm <plugin>` (repos: exists, plugconf: not exists, vim repos: exists) (A, B, C, E, F)
func TestVoltRmPoptOnePluginNoPlugconf(t *testing.T) {
	testutils.SetUpEnv(t)
	out, err := testutils.RunVolt("get", "tyru/caw.vim")
	testutils.SuccessExit(t, out, err)
	reposPath := "github.com/tyru/caw.vim"
	if err := os.Remove(pathutil.PlugconfOf(reposPath)); err != nil {
		t.Fatal("failed to remove plugconf: " + err.Error())
	}
	out, err = testutils.RunVolt("rm", "-p", "tyru/caw.vim")
	// (A, B)
	testutils.SuccessExit(t, out, err)

	// (C)
	reposDir := pathutil.FullReposPathOf(reposPath)
	if pathutil.Exists(reposDir) {
		t.Fatal("repos was not removed: " + reposDir)
	}

	// (E)
	vimReposDir := pathutil.PackReposPathOf(reposPath)
	if pathutil.Exists(vimReposDir) {
		t.Fatal("vim repos was not removed: " + vimReposDir)
	}

	// (F)
	testReposPathWereRemoved(t, reposPath)
}

// Run `volt rm -p <plugin>` (repos: not exists, plugconf: exists, vim repos: exists) (!A, B, D)
func TestVoltRmPoptOnePluginNoRepos(t *testing.T) {
	testutils.SetUpEnv(t)
	out, err := testutils.RunVolt("get", "tyru/caw.vim")
	testutils.SuccessExit(t, out, err)
	reposPath := "github.com/tyru/caw.vim"
	if err := os.RemoveAll(pathutil.FullReposPathOf(reposPath)); err != nil {
		t.Fatal("failed to remove repos: " + err.Error())
	}
	out, err = testutils.RunVolt("rm", "-p", "tyru/caw.vim")
	// (!A) "[WARN] no repository was installed: (voltpath)/repos/github.com/tyru/caw.vim"
	outstr := string(out)
	if !strings.Contains(outstr, "[WARN]") && !strings.Contains(outstr, "[ERROR]") {
		t.Fatal("expected error but no error: " + outstr)
	}
	// (B)
	if err != nil {
		t.Fatal("expected success exit but exited with failure: " + err.Error())
	}

	// (D)
	plugconf := pathutil.PlugconfOf(reposPath)
	if pathutil.Exists(plugconf) {
		t.Fatal("plugconf was not removed: " + plugconf)
	}
}

// Run `volt rm -p <plugin1> <plugin2>` (repos: exists, plugconf: exists, vim repos: exists) (A, B, C, D, E, F)
func TestVoltRmPoptTwoOrMorePluginNoPlugconf(t *testing.T) {
	testutils.SetUpEnv(t)
	out, err := testutils.RunVolt("get", "tyru/caw.vim", "tyru/capture.vim")
	testutils.SuccessExit(t, out, err)
	out, err = testutils.RunVolt("rm", "-p", "tyru/caw.vim", "tyru/capture.vim")
	// (A, B)
	testutils.SuccessExit(t, out, err)
	cawReposPath := "github.com/tyru/caw.vim"
	captureReposPath := "github.com/tyru/capture.vim"

	for _, reposPath := range []string{cawReposPath, captureReposPath} {
		// (C)
		reposDir := pathutil.FullReposPathOf(reposPath)
		if pathutil.Exists(reposDir) {
			t.Fatal("repos was not removed: " + reposDir)
		}

		// (D)
		plugconf := pathutil.PlugconfOf(reposPath)
		if pathutil.Exists(plugconf) {
			t.Fatal("plugconf was not removed: " + plugconf)
		}

		// (E)
		vimReposDir := pathutil.PackReposPathOf(reposPath)
		if pathutil.Exists(vimReposDir) {
			t.Fatal("vim repos was not removed: " + vimReposDir)
		}

		// (F)
		testReposPathWereRemoved(t, reposPath)
	}
}

// [error] Run `volt rm -p <plugin>` (repos: not exists, plugconf: not exists, vim repos: exists) (!A, !B)
func TestErrVoltRmPoptOnePluginNoReposNoPlugconf(t *testing.T) {
	testutils.SetUpEnv(t)
	out, err := testutils.RunVolt("get", "tyru/caw.vim")
	testutils.SuccessExit(t, out, err)
	reposPath := "github.com/tyru/caw.vim"
	if err := os.RemoveAll(pathutil.FullReposPathOf(reposPath)); err != nil {
		t.Fatal("failed to remove repos: " + err.Error())
	}
	if err := os.Remove(pathutil.PlugconfOf(reposPath)); err != nil {
		t.Fatal("failed to remove plugconf: " + err.Error())
	}
	out, err = testutils.RunVolt("rm", "-p", "tyru/caw.vim")
	// (!A, !B)
	testutils.FailExit(t, out, err)

	// TODO: Show error message that repos and vim repos are mismatch
}

func testReposPathWereRemoved(t *testing.T, reposPath string) {
	lockJSON, err := lockjson.Read()
	if err != nil {
		t.Fatal("lockjson.Read() returned non-nil error: " + err.Error())
	}
	if lockJSON.Repos.Contains(reposPath) {
		t.Fatal("repos was not removed from lock.json/repos: " + reposPath)
	}
	for i := range lockJSON.Profiles {
		if lockJSON.Profiles[i].ReposPath.Contains(reposPath) {
			t.Fatal("repos was not removed from lock.json/profiles/repos_path: " + reposPath)
		}
	}
}
