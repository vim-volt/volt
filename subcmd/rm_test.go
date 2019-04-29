package subcmd

import (
	"fmt"
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

// Run `volt rm <plugin>` (repos: exists, plugconf: exists, vim repos: exists) (A, B, !C, !D, E, F)
func TestVoltRmOnePlugin(t *testing.T) {
	testRmMatrix(t, func(t *testing.T, strategy string) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)
		testutil.InstallConfig(t, "strategy-"+strategy+".toml")

		out, err := testutil.RunVolt("get", "tyru/caw.vim")
		testutil.SuccessExit(t, out, err)

		// =============== run =============== //

		out, err = testutil.RunVolt("rm", "tyru/caw.vim")
		// (A, B)
		testutil.SuccessExit(t, out, err)
		reposPath := pathutil.ReposPath("github.com/tyru/caw.vim")

		// (!C)
		reposDir := reposPath.FullPath()
		if !pathutil.Exists(reposDir) {
			t.Error("repos was removed: " + reposDir)
		}

		// (!D)
		plugconf := reposPath.Plugconf()
		if !pathutil.Exists(plugconf) {
			t.Error("plugconf was removed: " + plugconf)
		}

		// (E)
		vimReposDir := reposPath.EncodeToPlugDirName()
		if pathutil.Exists(vimReposDir) {
			t.Error("vim repos was not removed: " + vimReposDir)
		}

		// (F)
		testReposPathWereRemoved(t, reposPath)
	})
}

// Run `volt rm -r <plugin>` (repos: exists, plugconf: exists, vim repos: exists) (A, B, C, !D, E, F)
func TestVoltRmRoptOnePlugin(t *testing.T) {
	testRmMatrix(t, func(t *testing.T, strategy string) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)
		testutil.InstallConfig(t, "strategy-"+strategy+".toml")

		out, err := testutil.RunVolt("get", "tyru/caw.vim")
		testutil.SuccessExit(t, out, err)

		// =============== run =============== //

		out, err = testutil.RunVolt("rm", "-r", "tyru/caw.vim")
		// (A, B)
		testutil.SuccessExit(t, out, err)
		reposPath := pathutil.ReposPath("github.com/tyru/caw.vim")

		// (C)
		reposDir := reposPath.FullPath()
		if pathutil.Exists(reposDir) {
			t.Error("repos was not removed: " + reposDir)
		}

		// (!D)
		plugconf := reposPath.Plugconf()
		if !pathutil.Exists(plugconf) {
			t.Error("plugconf was removed: " + plugconf)
		}

		// (E)
		vimReposDir := reposPath.EncodeToPlugDirName()
		if pathutil.Exists(vimReposDir) {
			t.Error("vim repos was not removed: " + vimReposDir)
		}

		// (F)
		testReposPathWereRemoved(t, reposPath)
	})
}

// Run `volt rm <plugin>` (repos: exists, plugconf: not exists, vim repos: exists) (A, B, !C, E, F)
func TestVoltRmOnePluginNoPlugconf(t *testing.T) {
	testRmMatrix(t, func(t *testing.T, strategy string) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)
		testutil.InstallConfig(t, "strategy-"+strategy+".toml")

		out, err := testutil.RunVolt("get", "tyru/caw.vim")
		testutil.SuccessExit(t, out, err)
		reposPath := pathutil.ReposPath("github.com/tyru/caw.vim")
		if err := os.Remove(reposPath.Plugconf()); err != nil {
			t.Error("failed to remove plugconf: " + err.Error())
		}

		// =============== run =============== //

		out, err = testutil.RunVolt("rm", "tyru/caw.vim")
		// (A, B)
		testutil.SuccessExit(t, out, err)

		// (!C)
		reposDir := reposPath.FullPath()
		if !pathutil.Exists(reposDir) {
			t.Error("repos was removed: " + reposDir)
		}

		// (E)
		vimReposDir := reposPath.EncodeToPlugDirName()
		if pathutil.Exists(vimReposDir) {
			t.Error("vim repos was not removed: " + vimReposDir)
		}

		// (F)
		testReposPathWereRemoved(t, reposPath)
	})
}

// Run `volt rm <plugin1> <plugin2>` (repos: exists, plugconf: exists, vim repos: exists) (A, B, !C, !D, E, F)
func TestVoltRmTwoOrMorePluginNoPlugconf(t *testing.T) {
	testRmMatrix(t, func(t *testing.T, strategy string) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)
		testutil.InstallConfig(t, "strategy-"+strategy+".toml")

		out, err := testutil.RunVolt("get", "tyru/caw.vim", "tyru/capture.vim")
		testutil.SuccessExit(t, out, err)

		// =============== run =============== //

		out, err = testutil.RunVolt("rm", "tyru/caw.vim", "tyru/capture.vim")
		// (A, B)
		testutil.SuccessExit(t, out, err)
		cawReposPath := pathutil.ReposPath("github.com/tyru/caw.vim")
		captureReposPath := pathutil.ReposPath("github.com/tyru/capture.vim")

		for _, reposPath := range []pathutil.ReposPath{cawReposPath, captureReposPath} {
			// (!C)
			reposDir := reposPath.FullPath()
			if !pathutil.Exists(reposDir) {
				t.Error("repos was removed: " + reposDir)
			}

			// (!D)
			plugconf := reposPath.Plugconf()
			if !pathutil.Exists(plugconf) {
				t.Error("plugconf was removed: " + plugconf)
			}

			// (E)
			vimReposDir := reposPath.EncodeToPlugDirName()
			if pathutil.Exists(vimReposDir) {
				t.Error("vim repos was not removed: " + vimReposDir)
			}

			// (F)
			testReposPathWereRemoved(t, reposPath)
		}
	})
}

// Run `volt rm <plugin>` (repos: not exists, plugconf: exists, vim repos: exists) (A, B, !D, E, F)
func TestVoltRmOnePluginNoRepos(t *testing.T) {
	testRmMatrix(t, func(t *testing.T, strategy string) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)
		testutil.InstallConfig(t, "strategy-"+strategy+".toml")

		out, err := testutil.RunVolt("get", "tyru/caw.vim")
		testutil.SuccessExit(t, out, err)
		reposPath := pathutil.ReposPath("github.com/tyru/caw.vim")
		if err := os.RemoveAll(reposPath.FullPath()); err != nil {
			t.Error("failed to remove repos: " + err.Error())
		}

		// =============== run =============== //

		out, err = testutil.RunVolt("rm", "tyru/caw.vim")
		// (A, B)
		testutil.SuccessExit(t, out, err)

		// (!D)
		plugconf := reposPath.Plugconf()
		if !pathutil.Exists(plugconf) {
			t.Error("plugconf was removed: " + plugconf)
		}

		// (E)
		vimReposDir := reposPath.EncodeToPlugDirName()
		if pathutil.Exists(vimReposDir) {
			t.Error("vim repos was not removed: " + vimReposDir)
		}

		// (F)
		testReposPathWereRemoved(t, reposPath)
	})
}

// Run `volt rm <plugin>` (repos: not exists, plugconf: not exists, vim repos: exists) (A, B, E, F)
func TestVoltRmOnePluginNoReposNoPlugconf(t *testing.T) {
	testRmMatrix(t, func(t *testing.T, strategy string) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)
		testutil.InstallConfig(t, "strategy-"+strategy+".toml")

		out, err := testutil.RunVolt("get", "tyru/caw.vim")
		testutil.SuccessExit(t, out, err)
		reposPath := pathutil.ReposPath("github.com/tyru/caw.vim")
		if err := os.RemoveAll(reposPath.FullPath()); err != nil {
			t.Error("failed to remove repos: " + err.Error())
		}
		if err := os.Remove(reposPath.Plugconf()); err != nil {
			t.Error("failed to remove plugconf: " + err.Error())
		}

		// =============== run =============== //

		out, err = testutil.RunVolt("rm", "tyru/caw.vim")
		// (A, B)
		testutil.SuccessExit(t, out, err)

		// (E)
		vimReposDir := reposPath.EncodeToPlugDirName()
		if pathutil.Exists(vimReposDir) {
			t.Error("vim repos was not removed: " + vimReposDir)
		}

		// (F)
		testReposPathWereRemoved(t, reposPath)
	})
}

// Run `volt rm -p <plugin>` (repos: exists, plugconf: exists, vim repos: exists) (A, B, !C, D, E, F)
func TestVoltRmPoptOnePlugin(t *testing.T) {
	testRmMatrix(t, func(t *testing.T, strategy string) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)
		testutil.InstallConfig(t, "strategy-"+strategy+".toml")

		out, err := testutil.RunVolt("get", "tyru/caw.vim")
		testutil.SuccessExit(t, out, err)

		// =============== run =============== //

		out, err = testutil.RunVolt("rm", "-p", "tyru/caw.vim")
		// (A, B)
		testutil.SuccessExit(t, out, err)
		reposPath := pathutil.ReposPath("github.com/tyru/caw.vim")

		// (!C)
		reposDir := reposPath.FullPath()
		if !pathutil.Exists(reposDir) {
			t.Error("repos was removed: " + reposDir)
		}

		// (D)
		plugconf := reposPath.Plugconf()
		if pathutil.Exists(plugconf) {
			t.Error("plugconf was not removed: " + plugconf)
		}

		// (E)
		vimReposDir := reposPath.EncodeToPlugDirName()
		if pathutil.Exists(vimReposDir) {
			t.Error("vim repos was not removed: " + vimReposDir)
		}

		// (F)
		testReposPathWereRemoved(t, reposPath)
	})
}

// Run `volt rm -p <plugin>` (repos: exists, plugconf: not exists, vim repos: exists) (A, B, !C, E, F)
func TestVoltRmPoptOnePluginNoPlugconf(t *testing.T) {
	testRmMatrix(t, func(t *testing.T, strategy string) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)
		testutil.InstallConfig(t, "strategy-"+strategy+".toml")

		out, err := testutil.RunVolt("get", "tyru/caw.vim")
		testutil.SuccessExit(t, out, err)
		reposPath := pathutil.ReposPath("github.com/tyru/caw.vim")
		if err := os.Remove(reposPath.Plugconf()); err != nil {
			t.Error("failed to remove plugconf: " + err.Error())
		}

		// =============== run =============== //

		out, err = testutil.RunVolt("rm", "-p", "tyru/caw.vim")
		// (A, B)
		testutil.SuccessExit(t, out, err)

		// (!C)
		reposDir := reposPath.FullPath()
		if !pathutil.Exists(reposDir) {
			t.Error("repos was removed: " + reposDir)
		}

		// (E)
		vimReposDir := reposPath.EncodeToPlugDirName()
		if pathutil.Exists(vimReposDir) {
			t.Error("vim repos was not removed: " + vimReposDir)
		}

		// (F)
		testReposPathWereRemoved(t, reposPath)
	})
}

// Run `volt rm -p <plugin>` (repos: not exists, plugconf: exists, vim repos: exists) (A, B, D, E, F)
func TestVoltRmPoptOnePluginNoRepos(t *testing.T) {
	testRmMatrix(t, func(t *testing.T, strategy string) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)
		testutil.InstallConfig(t, "strategy-"+strategy+".toml")

		out, err := testutil.RunVolt("get", "tyru/caw.vim")
		testutil.SuccessExit(t, out, err)
		reposPath := pathutil.ReposPath("github.com/tyru/caw.vim")
		if err := os.RemoveAll(reposPath.FullPath()); err != nil {
			t.Error("failed to remove repos: " + err.Error())
		}

		// =============== run =============== //

		out, err = testutil.RunVolt("rm", "-p", "tyru/caw.vim")
		// (A, B)
		testutil.SuccessExit(t, out, err)

		// (D)
		plugconf := reposPath.Plugconf()
		if pathutil.Exists(plugconf) {
			t.Error("plugconf was not removed: " + plugconf)
		}

		// (E)
		vimReposDir := reposPath.EncodeToPlugDirName()
		if pathutil.Exists(vimReposDir) {
			t.Error("vim repos was not removed: " + vimReposDir)
		}

		// (F)
		testReposPathWereRemoved(t, reposPath)
	})
}

// Run `volt rm -p <plugin1> <plugin2>` (repos: exists, plugconf: exists, vim repos: exists) (A, B, !C, D, E, F)
func TestVoltRmPoptTwoOrMorePluginNoPlugconf(t *testing.T) {
	testRmMatrix(t, func(t *testing.T, strategy string) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)
		testutil.InstallConfig(t, "strategy-"+strategy+".toml")

		out, err := testutil.RunVolt("get", "tyru/caw.vim", "tyru/capture.vim")
		testutil.SuccessExit(t, out, err)

		// =============== run =============== //

		out, err = testutil.RunVolt("rm", "-p", "tyru/caw.vim", "tyru/capture.vim")
		// (A, B)
		testutil.SuccessExit(t, out, err)
		cawReposPath := pathutil.ReposPath("github.com/tyru/caw.vim")
		captureReposPath := pathutil.ReposPath("github.com/tyru/capture.vim")

		for _, reposPath := range []pathutil.ReposPath{cawReposPath, captureReposPath} {
			// (!C)
			reposDir := reposPath.FullPath()
			if !pathutil.Exists(reposDir) {
				t.Error("repos was removed: " + reposDir)
			}

			// (D)
			plugconf := reposPath.Plugconf()
			if pathutil.Exists(plugconf) {
				t.Error("plugconf was not removed: " + plugconf)
			}

			// (E)
			vimReposDir := reposPath.EncodeToPlugDirName()
			if pathutil.Exists(vimReposDir) {
				t.Error("vim repos was not removed: " + vimReposDir)
			}

			// (F)
			testReposPathWereRemoved(t, reposPath)
		}
	})
}

// Run `volt rm -p <plugin>` (repos: not exists, plugconf: not exists, vim repos: exists) (A, B, E, F)
func TestVoltRmPoptOnePluginNoReposNoPlugconf(t *testing.T) {
	testRmMatrix(t, func(t *testing.T, strategy string) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)
		testutil.InstallConfig(t, "strategy-"+strategy+".toml")

		out, err := testutil.RunVolt("get", "tyru/caw.vim")
		testutil.SuccessExit(t, out, err)
		reposPath := pathutil.ReposPath("github.com/tyru/caw.vim")
		if err := os.RemoveAll(reposPath.FullPath()); err != nil {
			t.Error("failed to remove repos: " + err.Error())
		}
		if err := os.Remove(reposPath.Plugconf()); err != nil {
			t.Error("failed to remove plugconf: " + err.Error())
		}

		// =============== run =============== //

		out, err = testutil.RunVolt("rm", "-p", "tyru/caw.vim")
		// (A, B)
		testutil.SuccessExit(t, out, err)

		// (E)
		vimReposDir := reposPath.EncodeToPlugDirName()
		if pathutil.Exists(vimReposDir) {
			t.Error("vim repos was not removed: " + vimReposDir)
		}

		// (F)
		testReposPathWereRemoved(t, reposPath)
	})
}

// [error] Specify invalid argument (!A, !B)
func TestErrVoltRmInvalidArgs(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)
	defer testutil.CleanUpEnv(t)

	// =============== run =============== //

	out, err := testutil.RunVolt("rm", "caw.vim")
	// (!A, !B)
	testutil.FailExit(t, out, err)
}

// [error] Specify plugin which does not exist (!A, !B)
func TestErrVoltRmNotFound(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)
	defer testutil.CleanUpEnv(t)

	// =============== run =============== //

	out, err := testutil.RunVolt("rm", "vim-volt/not_found")
	// (!A, !B)
	testutil.FailExit(t, out, err)
}

func testReposPathWereRemoved(t *testing.T, reposPath pathutil.ReposPath) {
	t.Helper()
	lockJSON, err := lockjson.Read()
	if err != nil {
		t.Error("lockjson.Read() returned non-nil error: " + err.Error())
	}
	if lockJSON.Repos.Contains(reposPath) {
		t.Error("repos was not removed from lock.json/repos: " + reposPath)
	}
	for i := range lockJSON.Profiles {
		if lockJSON.Profiles[i].ReposPath.Contains(reposPath) {
			t.Error("repos was not removed from lock.json/profiles/repos_path: " + reposPath)
		}
	}
}

func testRmMatrix(t *testing.T, f func(*testing.T, string)) {
	for _, strategy := range testutil.AvailableStrategies() {
		t.Run(fmt.Sprintf("strategy=%v", strategy), func(t *testing.T) {
			f(t, strategy)
		})
	}
}
