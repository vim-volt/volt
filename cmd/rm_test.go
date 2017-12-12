package cmd

import (
	"os"
	"testing"

	"github.com/vim-volt/volt/internal/testutil"
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
	testutil.SetUpEnv(t)
	out, err := testutil.RunVolt("get", "tyru/caw.vim")
	testutil.SuccessExit(t, out, err)
	out, err = testutil.RunVolt("rm", "tyru/caw.vim")
	// (A, B)
	testutil.SuccessExit(t, out, err)
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
	testutil.SetUpEnv(t)
	out, err := testutil.RunVolt("get", "tyru/caw.vim")
	testutil.SuccessExit(t, out, err)
	reposPath := "github.com/tyru/caw.vim"
	if err := os.Remove(pathutil.PlugconfOf(reposPath)); err != nil {
		t.Fatal("failed to remove plugconf: " + err.Error())
	}
	out, err = testutil.RunVolt("rm", "tyru/caw.vim")
	// (A, B)
	testutil.SuccessExit(t, out, err)

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
	testutil.SetUpEnv(t)
	out, err := testutil.RunVolt("get", "tyru/caw.vim", "tyru/capture.vim")
	testutil.SuccessExit(t, out, err)
	out, err = testutil.RunVolt("rm", "tyru/caw.vim", "tyru/capture.vim")
	// (A, B)
	testutil.SuccessExit(t, out, err)
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

// Run `volt rm <plugin>` (repos: not exists, plugconf: exists, vim repos: exists) (A, B, !D, E, F)
func TestVoltRmOnePluginNoRepos(t *testing.T) {
	testutil.SetUpEnv(t)
	out, err := testutil.RunVolt("get", "tyru/caw.vim")
	testutil.SuccessExit(t, out, err)
	reposPath := "github.com/tyru/caw.vim"
	if err := os.RemoveAll(pathutil.FullReposPathOf(reposPath)); err != nil {
		t.Fatal("failed to remove repos: " + err.Error())
	}
	out, err = testutil.RunVolt("rm", "tyru/caw.vim")
	// (A, B)
	testutil.SuccessExit(t, out, err)

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

// Run `volt rm <plugin>` (repos: not exists, plugconf: not exists, vim repos: exists) (A, B, E, F)
func TestVoltRmOnePluginNoReposNoPlugconf(t *testing.T) {
	testutil.SetUpEnv(t)
	out, err := testutil.RunVolt("get", "tyru/caw.vim")
	testutil.SuccessExit(t, out, err)
	reposPath := "github.com/tyru/caw.vim"
	if err := os.RemoveAll(pathutil.FullReposPathOf(reposPath)); err != nil {
		t.Fatal("failed to remove repos: " + err.Error())
	}
	if err := os.Remove(pathutil.PlugconfOf(reposPath)); err != nil {
		t.Fatal("failed to remove plugconf: " + err.Error())
	}
	out, err = testutil.RunVolt("rm", "tyru/caw.vim")
	// (A, B)
	testutil.SuccessExit(t, out, err)

	// (E)
	vimReposDir := pathutil.PackReposPathOf(reposPath)
	if pathutil.Exists(vimReposDir) {
		t.Fatal("vim repos was not removed: " + vimReposDir)
	}

	// (F)
	testReposPathWereRemoved(t, reposPath)
}

// Run `volt rm -p <plugin>` (repos: exists, plugconf: exists, vim repos: exists) (A, B, C, D, E, F)
func TestVoltRmPoptOnePlugin(t *testing.T) {
	testutil.SetUpEnv(t)
	out, err := testutil.RunVolt("get", "tyru/caw.vim")
	testutil.SuccessExit(t, out, err)
	out, err = testutil.RunVolt("rm", "-p", "tyru/caw.vim")
	// (A, B)
	testutil.SuccessExit(t, out, err)
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

// Run `volt rm -p <plugin>` (repos: exists, plugconf: not exists, vim repos: exists) (A, B, C, E, F)
func TestVoltRmPoptOnePluginNoPlugconf(t *testing.T) {
	testutil.SetUpEnv(t)
	out, err := testutil.RunVolt("get", "tyru/caw.vim")
	testutil.SuccessExit(t, out, err)
	reposPath := "github.com/tyru/caw.vim"
	if err := os.Remove(pathutil.PlugconfOf(reposPath)); err != nil {
		t.Fatal("failed to remove plugconf: " + err.Error())
	}
	out, err = testutil.RunVolt("rm", "-p", "tyru/caw.vim")
	// (A, B)
	testutil.SuccessExit(t, out, err)

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

// Run `volt rm -p <plugin>` (repos: not exists, plugconf: exists, vim repos: exists) (A, B, D, E, F)
func TestVoltRmPoptOnePluginNoRepos(t *testing.T) {
	testutil.SetUpEnv(t)
	out, err := testutil.RunVolt("get", "tyru/caw.vim")
	testutil.SuccessExit(t, out, err)
	reposPath := "github.com/tyru/caw.vim"
	if err := os.RemoveAll(pathutil.FullReposPathOf(reposPath)); err != nil {
		t.Fatal("failed to remove repos: " + err.Error())
	}
	out, err = testutil.RunVolt("rm", "-p", "tyru/caw.vim")
	// (A, B)
	testutil.SuccessExit(t, out, err)

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

// Run `volt rm -p <plugin1> <plugin2>` (repos: exists, plugconf: exists, vim repos: exists) (A, B, C, D, E, F)
func TestVoltRmPoptTwoOrMorePluginNoPlugconf(t *testing.T) {
	testutil.SetUpEnv(t)
	out, err := testutil.RunVolt("get", "tyru/caw.vim", "tyru/capture.vim")
	testutil.SuccessExit(t, out, err)
	out, err = testutil.RunVolt("rm", "-p", "tyru/caw.vim", "tyru/capture.vim")
	// (A, B)
	testutil.SuccessExit(t, out, err)
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

// Run `volt rm -p <plugin>` (repos: not exists, plugconf: not exists, vim repos: exists) (A, B, E, F)
func TestVoltRmPoptOnePluginNoReposNoPlugconf(t *testing.T) {
	testutil.SetUpEnv(t)
	out, err := testutil.RunVolt("get", "tyru/caw.vim")
	testutil.SuccessExit(t, out, err)
	reposPath := "github.com/tyru/caw.vim"
	if err := os.RemoveAll(pathutil.FullReposPathOf(reposPath)); err != nil {
		t.Fatal("failed to remove repos: " + err.Error())
	}
	if err := os.Remove(pathutil.PlugconfOf(reposPath)); err != nil {
		t.Fatal("failed to remove plugconf: " + err.Error())
	}
	out, err = testutil.RunVolt("rm", "-p", "tyru/caw.vim")
	// (A, B)
	testutil.SuccessExit(t, out, err)

	// (E)
	vimReposDir := pathutil.PackReposPathOf(reposPath)
	if pathutil.Exists(vimReposDir) {
		t.Fatal("vim repos was not removed: " + vimReposDir)
	}

	// (F)
	testReposPathWereRemoved(t, reposPath)
}

// [error] Specify invalid argument (!A, !B)
func TestErrVoltRmInvalidArgs(t *testing.T) {
	testutil.SetUpEnv(t)
	out, err := testutil.RunVolt("rm", "caw.vim")
	// (!A, !B)
	testutil.FailExit(t, out, err)
}

// [error] Specify plugin which does not exist (!A, !B)
func TestErrVoltRmNotFound(t *testing.T) {
	testutil.SetUpEnv(t)
	out, err := testutil.RunVolt("rm", "vim-volt/not_found")
	// (!A, !B)
	testutil.FailExit(t, out, err)
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
