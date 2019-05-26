package subcmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/vim-volt/volt/config"
	"github.com/vim-volt/volt/internal/testutil"
	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/pathutil"
)

// Checks:
// (A) Does not show `[ERROR]`, `[WARN]` messages
// (B) Exit with zero status

// Checks:
// (a) Changes current profile
// (b) Plugins of specified profile are installed under vim dir
//
// * Run `volt profile set <profile>` (`<profile>` is not current profile) (A, B, a, b)
// * Run `volt profile set <profile>` (`<profile>` is current profile) (!A, !B, !a)
// * Run `volt profile set -n <profile>` (`<profile>` is not current profile and non-existing profile) (A, B, a)
func TestVoltProfileSet(t *testing.T) {
	t.Run("Run `volt profile set <profile>` (`<profile>` is not current profile)", func(t *testing.T) {
		testProfileMatrix(t, func(t *testing.T, strategy string) {
			// =============== setup =============== //

			testutil.SetUpEnv(t)
			defer testutil.CleanUpEnv(t)

			reposPathList := []pathutil.ReposPath{pathutil.ReposPath("github.com/tyru/caw.vim")}
			teardown := testutil.SetUpRepos(t, "caw.vim", lockjson.ReposGitType, reposPathList, strategy)
			defer teardown()
			testutil.InstallConfig(t, "strategy-"+strategy+".toml")

			out, err := testutil.RunVolt("profile", "new", "foo")
			testutil.SuccessExit(t, out, err)
			out, err = testutil.RunVolt("profile", "rm", "default", "github.com/tyru/caw.vim")
			testutil.SuccessExit(t, out, err)
			out, err = testutil.RunVolt("profile", "add", "foo", "github.com/tyru/caw.vim")
			testutil.SuccessExit(t, out, err)

			// =============== run =============== //

			profileName := "foo"
			out, err = testutil.RunVolt("profile", "set", profileName)
			// (A, B)
			testutil.SuccessExit(t, out, err)

			// (a)
			lockJSON, err := lockjson.Read()
			if err != nil {
				t.Error("lockjson.Read() returned non-nil error: " + err.Error())
			}
			if lockJSON.CurrentProfileName != profileName {
				t.Errorf("expected: %s, got: %s", profileName, lockJSON.CurrentProfileName)
			}

			// (b)
			for _, reposPath := range reposPathList {
				vimReposDir := reposPath.EncodeToPlugDirName()
				if !pathutil.Exists(vimReposDir) {
					t.Error("vim repos does not exist: " + vimReposDir)
				}
			}
		})
	})

	t.Run("Run `volt profile set <profile>` (`<profile>` is current profile)", func(t *testing.T) {
		testProfileMatrix(t, func(t *testing.T, strategy string) {
			// =============== setup =============== //

			testutil.SetUpEnv(t)
			defer testutil.CleanUpEnv(t)
			testutil.InstallConfig(t, "strategy-"+strategy+".toml")

			// =============== run =============== //

			profileName := "default"
			out, err := testutil.RunVolt("profile", "set", profileName)
			// (!A, !B)
			testutil.FailExit(t, out, err)

			// (!a)
			lockJSON, err := lockjson.Read()
			if err != nil {
				t.Error("lockjson.Read() returned non-nil error: " + err.Error())
			}
			if lockJSON.CurrentProfileName != profileName {
				t.Errorf("expected: %s, got: %s", profileName, lockJSON.CurrentProfileName)
			}
		})
	})

	t.Run("Run `volt profile set -n <profile>` (`<profile>` is not current profile and non-existing profile)", func(t *testing.T) {
		testProfileMatrix(t, func(t *testing.T, strategy string) {
			// =============== setup =============== //

			testutil.SetUpEnv(t)
			defer testutil.CleanUpEnv(t)
			testutil.InstallConfig(t, "strategy-"+strategy+".toml")

			// =============== run =============== //

			profileName := "bar"
			out, err := testutil.RunVolt("profile", "set", "-n", profileName)
			// (A, B)
			testutil.SuccessExit(t, out, err)

			// (a)
			lockJSON, err := lockjson.Read()
			if err != nil {
				t.Error("lockjson.Read() returned non-nil error: " + err.Error())
			}
			if lockJSON.CurrentProfileName != profileName {
				t.Errorf("expected: %s, got: %s", profileName, lockJSON.CurrentProfileName)
			}
		})
	})
}

// Checks:
// (a) Output has profile name
// (b) Output has "repos path"
//
// * Run `volt profile show <profile>` (`<profile>` is existing profile) (A, B, a, b)
// * Run `volt profile show -current` (A, B, a, b)
// * Run `volt profile show <profile>` (`<profile>` is non-existing profile) (!A, !B, !a, !b)
func TestVoltProfileShow(t *testing.T) {
	t.Run("Run `volt profile show <profile>` (`<profile>` is existing profile)", func(t *testing.T) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)

		// =============== run =============== //

		out, err := testutil.RunVolt("profile", "show", "default")
		// (A, B)
		testutil.SuccessExit(t, out, err)

		// (a, b)
		outstr := string(out)
		if !strings.Contains(outstr, "name: default\n") {
			t.Errorf("Expected 'name: default' line, but got: %s", outstr)
		}
		if !strings.Contains(outstr, "repos path:\n") {
			t.Errorf("Expected 'repos path:' line, but got: %s", outstr)
		}
	})

	t.Run("Run `volt profile show -current`", func(t *testing.T) {
		out, err := testutil.RunVolt("profile", "show", "-current")
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)

		// =============== run =============== //

		// (A, B)
		testutil.SuccessExit(t, out, err)

		// (a, b)
		outstr := string(out)
		if !strings.Contains(outstr, "name: default\n") {
			t.Errorf("Expected 'name: default' line, but got: %s", outstr)
		}
		if !strings.Contains(outstr, "repos path:\n") {
			t.Errorf("Expected 'repos path:' line, but got: %s", outstr)
		}
	})

	t.Run("Run `volt profile show <profile>` (`<profile>` is non-existing profile)", func(t *testing.T) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)

		// =============== run =============== //

		out, err := testutil.RunVolt("profile", "show", "bar")
		// (!A, !B)
		testutil.FailExit(t, out, err)

		// (!a, !b)
		outstr := string(out)
		expected := "[ERROR] profile 'bar' does not exist"
		if strings.Trim(outstr, " \t\r\n") != expected {
			t.Errorf("Expected '%s' line, but got: '%s'", expected, outstr)
		}
	})
}

// Checks:
// (a) Current profile is marked as "*"
// (b) Created profile are showed
//
// * Run `volt profile list` (A, B, a)
// * Run `volt profile list` after creating profile (A, B, a, b)
func TestVoltProfileList(t *testing.T) {
	t.Run("Run `volt profile list`", func(t *testing.T) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)

		// =============== run =============== //

		out, err := testutil.RunVolt("profile", "list")
		// (A, B)
		testutil.SuccessExit(t, out, err)
		// (a, b)
		outstr := strings.Trim(string(out), " \t\r\n")
		expected := "* default"
		if outstr != expected {
			t.Errorf("Expected '%s' output, but got: '%s'", expected, outstr)
		}
	})

	t.Run("Run `volt profile list` after creating profile", func(t *testing.T) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)
		out, err := testutil.RunVolt("profile", "new", "foo")
		testutil.SuccessExit(t, out, err)

		// =============== run =============== //

		out, err = testutil.RunVolt("profile", "list")
		// (A, B)
		testutil.SuccessExit(t, out, err)
		// (a, b)
		outstr := strings.Trim(string(out), " \t\r\n")
		expected := "* default\n  foo"
		if outstr != expected {
			t.Errorf("Expected '%s' output, but got: '%s'", expected, outstr)
		}
	})
}

// Checks:
// (a) Created profile exists
// (b) Current profile is not changed
//
// * Run `volt profile new <profile>` (<profile> is not current profile and non-existing profile) (A, B, a, b)
// * Run `volt profile new <profile>` (<profile> is current profile) (!A, !B, a, b)
// * Run `volt profile new <profile>` (<profile> is existing profile) (!A, !B, a, b)
func TestVoltProfileNew(t *testing.T) {
	t.Run("Run `volt profile new <profile>` (<profile> is not current profile and non-existing profile)", func(t *testing.T) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)

		// =============== run =============== //

		out, err := testutil.RunVolt("profile", "new", "foo")
		// (A, B)
		testutil.SuccessExit(t, out, err)

		lockJSON, err := lockjson.Read()
		if err != nil {
			t.Error("lockjson.Read() returned non-nil error: " + err.Error())
		}
		// (a)
		if lockJSON.Profiles.FindIndexByName("foo") == -1 {
			t.Errorf("expected profile '%s' exists, but does not exist", "foo")
		}
		// (b)
		if lockJSON.CurrentProfileName != "default" {
			t.Errorf("expected current profile is '%s', but got: %s", "default", lockJSON.CurrentProfileName)
		}
	})

	t.Run("Run `volt profile new <profile>` (<profile> is current profile)", func(t *testing.T) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)

		// =============== run =============== //

		out, err := testutil.RunVolt("profile", "new", "default")
		// (!A, !B)
		testutil.FailExit(t, out, err)

		lockJSON, err := lockjson.Read()
		if err != nil {
			t.Error("lockjson.Read() returned non-nil error: " + err.Error())
		}
		// (a)
		if lockJSON.Profiles.FindIndexByName("default") == -1 {
			t.Errorf("expected profile '%s' exists, but does not exist", "default")
		}
		// (b)
		if lockJSON.CurrentProfileName != "default" {
			t.Errorf("expected current profile is '%s', but got: %s", "default", lockJSON.CurrentProfileName)
		}
	})

	t.Run("Run `volt profile new <profile>` (<profile> is existing profile)", func(t *testing.T) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)
		out, err := testutil.RunVolt("profile", "new", "bar")
		testutil.SuccessExit(t, out, err)

		// =============== run =============== //

		out, err = testutil.RunVolt("profile", "new", "bar")
		// (!A, !B)
		testutil.FailExit(t, out, err)

		lockJSON, err := lockjson.Read()
		if err != nil {
			t.Error("lockjson.Read() returned non-nil error: " + err.Error())
		}
		// (a)
		if lockJSON.Profiles.FindIndexByName("bar") == -1 {
			t.Errorf("expected profile '%s' exists, but does not exist", "bar")
		}
		// (b)
		if lockJSON.CurrentProfileName != "default" {
			t.Errorf("expected current profile is '%s', but got: %s", "default", lockJSON.CurrentProfileName)
		}
	})
}

// Checks:
// (a) profile is removed
//
// * Run `volt profile destroy <profile>` (<profile> is not current profile and existing profile) (A, B, a)
// * Run `volt profile destroy <profile>` (<profile> is current profile) (!A, !B, !a)
// * Run `volt profile destroy <profile>` (<profile> is non-existing profile) (!A, !B)
func TestVoltProfileDestroy(t *testing.T) {
	t.Run("Run `volt profile destroy <profile>` (<profile> is not current profile and existing profile)", func(t *testing.T) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)
		out, err := testutil.RunVolt("profile", "new", "foo")
		// (A, B)
		testutil.SuccessExit(t, out, err)

		// =============== run =============== //

		out, err = testutil.RunVolt("profile", "destroy", "foo")
		// (A, B)
		testutil.SuccessExit(t, out, err)

		lockJSON, err := lockjson.Read()
		if err != nil {
			t.Error("lockjson.Read() returned non-nil error: " + err.Error())
		}
		// (a)
		if lockJSON.Profiles.FindIndexByName("foo") != -1 {
			t.Errorf("expected profile '%s' does not exist, but does exist", "foo")
		}
	})

	t.Run("Run `volt profile destroy <profile>` (<profile> is current profile)", func(t *testing.T) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)

		// =============== run =============== //

		out, err := testutil.RunVolt("profile", "destroy", "default")
		// (!A, !B)
		testutil.FailExit(t, out, err)

		lockJSON, err := lockjson.Read()
		if err != nil {
			t.Error("lockjson.Read() returned non-nil error: " + err.Error())
		}
		// (!a)
		if lockJSON.Profiles.FindIndexByName("default") == -1 {
			t.Errorf("expected profile '%s' does exist, but does not exist", "default")
		}
	})

	t.Run("Run `volt profile destroy <profile>` (<profile> is non-existing profile)", func(t *testing.T) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)

		// =============== run =============== //

		out, err := testutil.RunVolt("profile", "destroy", "foo")
		// (!A, !B)
		testutil.FailExit(t, out, err)
	})
}

// Checks:
// (a) source profile exists
// (b) destination profile exists
// (c) current profile was changed
// (d) current profile was changed to destination profile
//
// * Run `volt profile rename <src> <dst>` (<src>: exists & not current profile, <dst>: not exist) (A, B, !a, b, !c)
// * Run `volt profile rename <src> <dst>` (<src>: exists & current profile, <dst>: not exist) (A, B, !a, b, c, d)
// * Run `volt profile rename <src> <dst>` (<src>: exists & not current profile, <dst>: exists) (!A, !B, a, b, !c)
// * Run `volt profile rename <src> <dst>` (<src>: exists & current profile, <dst>: exists) (!A, !B, a, b, !c)
// * Run `volt profile rename <src> <dst>` (<src>: not exist, <dst>: exists) (!A, !B, !a, b, !c)
// * Run `volt profile rename <src> <dst>` (<src>: not exist, <dst>: not exist) (!A, !B, !a, !b, !c)
func TestVoltProfileRename(t *testing.T) {
	t.Run("Run `volt profile rename <src> <dst>` (<src>: exists & not current profile, <dst>: not exist)", func(t *testing.T) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)
		out, err := testutil.RunVolt("profile", "set", "-n", "foo")
		testutil.SuccessExit(t, out, err)

		oldLockJSON, err := lockjson.Read()
		if err != nil {
			t.Error("lockjson.Read() returned non-nil error: " + err.Error())
		}

		// =============== run =============== //

		src, dst := "default", "bar"
		out, err = testutil.RunVolt("profile", "rename", src, dst)
		// (A, B)
		testutil.SuccessExit(t, out, err)

		lockJSON, err := lockjson.Read()
		if err != nil {
			t.Error("lockjson.Read() returned non-nil error: " + err.Error())
		}

		// (!a)
		if lockJSON.Profiles.FindIndexByName(src) != -1 {
			t.Errorf("expected profile '%s' does not exist, but does exist", src)
		}
		// (b)
		if lockJSON.Profiles.FindIndexByName(dst) == -1 {
			t.Errorf("expected profile '%s' does exist, but does not exist", dst)
		}
		// (!c)
		if lockJSON.CurrentProfileName != oldLockJSON.CurrentProfileName {
			t.Errorf("expected current profile did not change but changed: %s -> %s", oldLockJSON.CurrentProfileName, lockJSON.CurrentProfileName)
		}
	})

	t.Run("Run `volt profile rename <src> <dst>` (<src>: exists & current profile, <dst>: not exist)", func(t *testing.T) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)
		oldLockJSON, err := lockjson.Read()
		if err != nil {
			t.Error("lockjson.Read() returned non-nil error: " + err.Error())
		}

		// =============== run =============== //

		src, dst := "default", "foo"
		out, err := testutil.RunVolt("profile", "rename", src, dst)
		// (A, B)
		testutil.SuccessExit(t, out, err)

		lockJSON, err := lockjson.Read()
		if err != nil {
			t.Error("lockjson.Read() returned non-nil error: " + err.Error())
		}

		// (!a)
		if lockJSON.Profiles.FindIndexByName(src) != -1 {
			t.Errorf("expected profile '%s' does not exist, but does exist", src)
		}
		// (b)
		if lockJSON.Profiles.FindIndexByName(dst) == -1 {
			t.Errorf("expected profile '%s' does exist, but does not exist", dst)
		}
		// (c)
		if lockJSON.CurrentProfileName == oldLockJSON.CurrentProfileName {
			t.Errorf("expected current profile changed but did not change: %s -> %s", oldLockJSON.CurrentProfileName, lockJSON.CurrentProfileName)
		}
		// (d)
		if lockJSON.CurrentProfileName != dst {
			t.Errorf("expected current profile was changed to %q but got: %q", dst, lockJSON.CurrentProfileName)
		}
	})

	t.Run("Run `volt profile rename <src> <dst>` (<src>: exists & not current profile, <dst>: exists)", func(t *testing.T) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)
		out, err := testutil.RunVolt("profile", "new", "foo")
		testutil.SuccessExit(t, out, err)
		out, err = testutil.RunVolt("profile", "new", "bar")
		testutil.SuccessExit(t, out, err)

		oldLockJSON, err := lockjson.Read()
		if err != nil {
			t.Error("lockjson.Read() returned non-nil error: " + err.Error())
		}

		// =============== run =============== //

		src, dst := "foo", "bar"
		out, err = testutil.RunVolt("profile", "rename", src, dst)
		// (!A, !B)
		testutil.FailExit(t, out, err)

		lockJSON, err := lockjson.Read()
		if err != nil {
			t.Error("lockjson.Read() returned non-nil error: " + err.Error())
		}

		// (a)
		if lockJSON.Profiles.FindIndexByName(src) == -1 {
			t.Errorf("expected profile '%s' does exist, but does not exist", src)
		}
		// (b)
		if lockJSON.Profiles.FindIndexByName(dst) == -1 {
			t.Errorf("expected profile '%s' does exist, but does not exist", dst)
		}
		// (!c)
		if lockJSON.CurrentProfileName != oldLockJSON.CurrentProfileName {
			t.Errorf("expected current profile did not change but changed: %s -> %s", oldLockJSON.CurrentProfileName, lockJSON.CurrentProfileName)
		}
	})

	t.Run("Run `volt profile rename <src> <dst>` (<src>: exists & current profile, <dst>: exists)", func(t *testing.T) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)
		out, err := testutil.RunVolt("profile", "new", "foo")
		testutil.SuccessExit(t, out, err)

		oldLockJSON, err := lockjson.Read()
		if err != nil {
			t.Error("lockjson.Read() returned non-nil error: " + err.Error())
		}

		// =============== run =============== //

		src, dst := "default", "foo"
		out, err = testutil.RunVolt("profile", "rename", src, dst)
		// (!A, !B)
		testutil.FailExit(t, out, err)

		lockJSON, err := lockjson.Read()
		if err != nil {
			t.Error("lockjson.Read() returned non-nil error: " + err.Error())
		}

		// (a)
		if lockJSON.Profiles.FindIndexByName(src) == -1 {
			t.Errorf("expected profile '%s' does exist, but does not exist", src)
		}
		// (b)
		if lockJSON.Profiles.FindIndexByName(dst) == -1 {
			t.Errorf("expected profile '%s' does exist, but does not exist", dst)
		}
		// (!c)
		if lockJSON.CurrentProfileName != oldLockJSON.CurrentProfileName {
			t.Errorf("expected current profile did not change but changed: %s -> %s", oldLockJSON.CurrentProfileName, lockJSON.CurrentProfileName)
		}
	})

	t.Run("Run `volt profile rename <src> <dst>` (<src>: not exist, <dst>: exists)", func(t *testing.T) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)
		out, err := testutil.RunVolt("profile", "new", "foo")
		testutil.SuccessExit(t, out, err)

		oldLockJSON, err := lockjson.Read()
		if err != nil {
			t.Error("lockjson.Read() returned non-nil error: " + err.Error())
		}

		// =============== run =============== //

		src, dst := "bar", "foo"
		out, err = testutil.RunVolt("profile", "rename", src, dst)
		// (!A, !B)
		testutil.FailExit(t, out, err)

		lockJSON, err := lockjson.Read()
		if err != nil {
			t.Error("lockjson.Read() returned non-nil error: " + err.Error())
		}

		// (!a)
		if lockJSON.Profiles.FindIndexByName(src) != -1 {
			t.Errorf("expected profile '%s' does not exist, but does exist", src)
		}
		// (b)
		if lockJSON.Profiles.FindIndexByName(dst) == -1 {
			t.Errorf("expected profile '%s' does exist, but does not exist", dst)
		}
		// (!c)
		if lockJSON.CurrentProfileName != oldLockJSON.CurrentProfileName {
			t.Errorf("expected current profile did not change but changed: %s -> %s", oldLockJSON.CurrentProfileName, lockJSON.CurrentProfileName)
		}
	})

	t.Run("Run `volt profile rename <src> <dst>` (<src>: not exist, <dst>: not exist)", func(t *testing.T) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)

		oldLockJSON, err := lockjson.Read()
		if err != nil {
			t.Error("lockjson.Read() returned non-nil error: " + err.Error())
		}

		// =============== run =============== //

		src, dst := "foo", "bar"
		out, err := testutil.RunVolt("profile", "rename", src, dst)
		// (!A, !B)
		testutil.FailExit(t, out, err)

		lockJSON, err := lockjson.Read()
		if err != nil {
			t.Error("lockjson.Read() returned non-nil error: " + err.Error())
		}

		// (!a)
		if lockJSON.Profiles.FindIndexByName(src) != -1 {
			t.Errorf("expected profile '%s' does not exist, but does exist", src)
		}
		// (!b)
		if lockJSON.Profiles.FindIndexByName(dst) != -1 {
			t.Errorf("expected profile '%s' does not exist, but does exist", dst)
		}
		// (!c)
		if lockJSON.CurrentProfileName != oldLockJSON.CurrentProfileName {
			t.Errorf("expected current profile did not change but changed: %s -> %s", oldLockJSON.CurrentProfileName, lockJSON.CurrentProfileName)
		}
	})
}

// Checks:
// (a) given repositories are added to profile
// (b) other profiles which was not specified do not change
// (c) specified profile exists
//
// * Run `volt profile add <profile> <repos>` (<profile>: exists, <repos>: exists) (A, B, a, b, c)
// * Run `volt profile add <profile> <repos1> <repos2>` (<profile>: exists, <repos1>,<repos2>: exists) (A, B, a, b, c)
// * Run `volt profile add -current <repos>` (<repos>: exists) (A, B, a, b, c)
// * Run `volt profile add <profile> <repos>` (<profile>: not exist, <repos>: exists) (!A, !B, b, !c)
// * Run `volt profile add <profile> <repos>` (<profile>: exists, <repos>: not exist) (!A, !B, !a, b, c)
// * Run `volt profile add <profile> <repos>` (<profile>: not exist, <repos>: not exist) (!A, !B, b, !c)
func TestVoltProfileAdd(t *testing.T) {
	t.Run("Run `volt profile add <profile> <repos>` (<profile>: exists, <repos>: exists)", func(t *testing.T) {
		testProfileMatrix(t, func(t *testing.T, strategy string) {
			// =============== setup =============== //

			testutil.SetUpEnv(t)
			defer testutil.CleanUpEnv(t)

			reposPathList := []pathutil.ReposPath{pathutil.ReposPath("github.com/tyru/caw.vim")}
			teardown := testutil.SetUpRepos(t, "caw.vim", lockjson.ReposGitType, reposPathList, config.SymlinkBuilder)
			defer teardown()
			testutil.InstallConfig(t, "strategy-"+strategy+".toml")

			out, err := testutil.RunVolt("profile", "new", "empty")
			testutil.SuccessExit(t, out, err)

			oldLockJSON, err := lockjson.Read()
			if err != nil {
				t.Error("lockjson.Read() returned non-nil error: " + err.Error())
			}

			// =============== run =============== //

			reposPath := pathutil.ReposPath("github.com/tyru/caw.vim")
			out, err = testutil.RunVolt("profile", "add", "empty", reposPath.String())
			// (A, B)
			testutil.SuccessExit(t, out, err)

			lockJSON, err := lockjson.Read()
			if err != nil {
				t.Error("lockjson.Read() returned non-nil error: " + err.Error())
			}
			reposList := getReposList(t, lockJSON, "empty")

			// (a)
			if !reposList.Contains(reposPath) {
				t.Errorf("expected '%s' is added to profile '%s', but not added", reposPath, "empty")
			}
			// (b)
			testNotChangedProfileExcept(t, oldLockJSON, lockJSON, "empty")
			// (c)
			if lockJSON.Profiles.FindIndexByName("empty") == -1 {
				t.Errorf("expected profile '%s' does exist, but does not exist", "empty")
			}
		})
	})

	t.Run("Run `volt profile add <profile> <repos1> <repos2>` (<profile>: exists, <repos1>,<repos2>: exists)", func(t *testing.T) {
		testProfileMatrix(t, func(t *testing.T, strategy string) {
			// =============== setup =============== //

			testutil.SetUpEnv(t)
			defer testutil.CleanUpEnv(t)

			reposPathList := pathutil.ReposPathList{pathutil.ReposPath("github.com/tyru/caw.vim"), pathutil.ReposPath("github.com/tyru/capture.vim")}
			teardown := testutil.SetUpRepos(t, "caw-and-capture", lockjson.ReposGitType, reposPathList, config.SymlinkBuilder)
			defer teardown()
			testutil.InstallConfig(t, "strategy-"+strategy+".toml")

			out, err := testutil.RunVolt("profile", "new", "empty")
			testutil.SuccessExit(t, out, err)

			oldLockJSON, err := lockjson.Read()
			if err != nil {
				t.Error("lockjson.Read() returned non-nil error: " + err.Error())
			}

			// =============== run =============== //

			args := append([]string{"profile", "add", "empty"}, reposPathList.Strings()...)
			out, err = testutil.RunVolt(args...)
			// (A, B)
			testutil.SuccessExit(t, out, err)

			lockJSON, err := lockjson.Read()
			if err != nil {
				t.Error("lockjson.Read() returned non-nil error: " + err.Error())
			}
			reposList := getReposList(t, lockJSON, "empty")

			// (a)
			for _, reposPath := range reposPathList {
				if !reposList.Contains(reposPath) {
					t.Errorf("expected '%s' is added to profile '%s', but not added", reposPath.String(), "empty")
				}
			}
			// (b)
			testNotChangedProfileExcept(t, oldLockJSON, lockJSON, "empty")
			// (c)
			if lockJSON.Profiles.FindIndexByName("empty") == -1 {
				t.Errorf("expected profile '%s' does exist, but does not exist", "empty")
			}
		})
	})

	t.Run("Run `volt profile add -current <repos>` (<repos>: exists)", func(t *testing.T) {
		testProfileMatrix(t, func(t *testing.T, strategy string) {
			// =============== setup =============== //

			testutil.SetUpEnv(t)
			defer testutil.CleanUpEnv(t)

			reposPathList := pathutil.ReposPathList{pathutil.ReposPath("github.com/tyru/caw.vim")}
			teardown := testutil.SetUpRepos(t, "caw.vim", lockjson.ReposGitType, reposPathList, config.SymlinkBuilder)
			defer teardown()
			testutil.InstallConfig(t, "strategy-"+strategy+".toml")

			out, err := testutil.RunVolt("profile", "set", "-n", "empty")
			testutil.SuccessExit(t, out, err)

			oldLockJSON, err := lockjson.Read()
			if err != nil {
				t.Error("lockjson.Read() returned non-nil error: " + err.Error())
			}

			// =============== run =============== //

			reposPath := pathutil.ReposPath("github.com/tyru/caw.vim")
			out, err = testutil.RunVolt("profile", "add", "-current", reposPath.String())
			// (A, B)
			testutil.SuccessExit(t, out, err)

			lockJSON, err := lockjson.Read()
			if err != nil {
				t.Error("lockjson.Read() returned non-nil error: " + err.Error())
			}
			reposList := getReposList(t, lockJSON, lockJSON.CurrentProfileName)

			// (a)
			if !reposList.Contains(reposPath) {
				t.Errorf("expected '%s' is added to profile '%s', but not added", reposPath, lockJSON.CurrentProfileName)
			}
			// (b)
			testNotChangedProfileExcept(t, oldLockJSON, lockJSON, lockJSON.CurrentProfileName)
			// (c)
			if lockJSON.Profiles.FindIndexByName(lockJSON.CurrentProfileName) == -1 {
				t.Errorf("expected profile '%s' does exist, but does not exist", lockJSON.CurrentProfileName)
			}
		})
	})

	t.Run("Run `volt profile add <profile> <repos>` (<profile>: not exist, <repos>: exists)", func(t *testing.T) {
		testProfileMatrix(t, func(t *testing.T, strategy string) {
			// =============== setup =============== //

			testutil.SetUpEnv(t)
			defer testutil.CleanUpEnv(t)

			reposPathList := pathutil.ReposPathList{pathutil.ReposPath("github.com/tyru/caw.vim")}
			teardown := testutil.SetUpRepos(t, "caw.vim", lockjson.ReposGitType, reposPathList, config.SymlinkBuilder)
			defer teardown()
			testutil.InstallConfig(t, "strategy-"+strategy+".toml")

			oldLockJSON, err := lockjson.Read()
			if err != nil {
				t.Error("lockjson.Read() returned non-nil error: " + err.Error())
			}

			// =============== run =============== //

			reposPath := "github.com/tyru/caw.vim"
			out, err := testutil.RunVolt("profile", "add", "not_existing_profile", reposPath)
			// (!A, !B)
			testutil.FailExit(t, out, err)

			lockJSON, err := lockjson.Read()
			if err != nil {
				t.Error("lockjson.Read() returned non-nil error: " + err.Error())
			}

			// (b)
			testNotChangedProfileExcept(t, oldLockJSON, lockJSON, "not_existing_profile")
			// (!c)
			if lockJSON.Profiles.FindIndexByName("not_existing_profile") != -1 {
				t.Errorf("expected profile '%s' does not exist, but does exist", "not_existing_profile")
			}
		})
	})

	t.Run("Run `volt profile add <profile> <repos>` (<profile>: exists, <repos>: not exist)", func(t *testing.T) {
		testProfileMatrix(t, func(t *testing.T, strategy string) {
			// =============== setup =============== //

			testutil.SetUpEnv(t)
			defer testutil.CleanUpEnv(t)
			testutil.InstallConfig(t, "strategy-"+strategy+".toml")

			out, err := testutil.RunVolt("profile", "new", "empty")
			testutil.SuccessExit(t, out, err)

			oldLockJSON, err := lockjson.Read()
			if err != nil {
				t.Error("lockjson.Read() returned non-nil error: " + err.Error())
			}

			// =============== run =============== //

			reposPath := pathutil.ReposPath("github.com/tyru/caw.vim")
			out, err = testutil.RunVolt("profile", "add", "empty", reposPath.String())
			// (!A, !B)
			testutil.FailExit(t, out, err)

			lockJSON, err := lockjson.Read()
			if err != nil {
				t.Error("lockjson.Read() returned non-nil error: " + err.Error())
			}
			reposList := getReposList(t, lockJSON, "empty")

			// (!a)
			if reposList.Contains(reposPath) {
				t.Errorf("expected '%s' is not added to profile '%s', but added", reposPath.String(), "empty")
			}
			// (b)
			testNotChangedProfileExcept(t, oldLockJSON, lockJSON, "empty")
			// (c)
			if lockJSON.Profiles.FindIndexByName("empty") == -1 {
				t.Errorf("expected profile '%s' does exist, but does not exist", "empty")
			}
		})
	})

	t.Run("Run `volt profile add <profile> <repos>` (<profile>: not exist, <repos>: not exist)", func(t *testing.T) {
		testProfileMatrix(t, func(t *testing.T, strategy string) {
			// =============== setup =============== //

			testutil.SetUpEnv(t)
			defer testutil.CleanUpEnv(t)
			testutil.InstallConfig(t, "strategy-"+strategy+".toml")

			oldLockJSON, err := lockjson.Read()
			if err != nil {
				t.Error("lockjson.Read() returned non-nil error: " + err.Error())
			}

			// =============== run =============== //

			reposPath := pathutil.ReposPath("github.com/tyru/caw.vim")
			out, err := testutil.RunVolt("profile", "add", "not_existing_profile", reposPath.String())
			// (!A, !B)
			testutil.FailExit(t, out, err)

			lockJSON, err := lockjson.Read()
			if err != nil {
				t.Error("lockjson.Read() returned non-nil error: " + err.Error())
			}

			// (b)
			testNotChangedProfileExcept(t, oldLockJSON, lockJSON, "not_existing_profile")
			// (!c)
			if lockJSON.Profiles.FindIndexByName("not_existing_profile") != -1 {
				t.Errorf("expected profile '%s' does not exist, but does exist", "empty")
			}
		})
	})
}

// Checks:
// (a) given repositories are removed from profile
// (b) other profiles which was not specified do not change
// (c) specified profile exists
//
// * Run `volt profile rm <profile> <repos>` (<profile>: exists, <repos>: exists) (A, B, a, b, c)
// * Run `volt profile rm <profile> <repos1> <repos2>` (<profile>: exists, <repos1>,<repos2>: exists) (A, B, a, b, c)
// * Run `volt profile rm -current <repos>` (<repos>: exists) (A, B, a, b, c)
// * Run `volt profile rm <profile> <repos>` (<profile>: not exist, <repos>: exists) (!A, !B, b, !c)
// * Run `volt profile rm <profile> <repos>` (<profile>: exists, <repos>: not exist) (!A, !B, a, b, c)
// * Run `volt profile rm <profile> <repos>` (<profile>: not exist, <repos>: not exist) (!A, !B, b, !c)
func TestVoltProfileRm(t *testing.T) {
	t.Run("Run `volt profile rm <profile> <repos>` (<profile>: exists, <repos>: exists)", func(t *testing.T) {
		testProfileMatrix(t, func(t *testing.T, strategy string) {
			// =============== setup =============== //

			testutil.SetUpEnv(t)
			defer testutil.CleanUpEnv(t)

			reposPathList := pathutil.ReposPathList{pathutil.ReposPath("github.com/tyru/caw.vim")}
			teardown := testutil.SetUpRepos(t, "caw.vim", lockjson.ReposGitType, reposPathList, config.SymlinkBuilder)
			defer teardown()
			testutil.InstallConfig(t, "strategy-"+strategy+".toml")

			oldLockJSON, err := lockjson.Read()
			if err != nil {
				t.Error("lockjson.Read() returned non-nil error: " + err.Error())
			}

			// =============== run =============== //

			reposPath := pathutil.ReposPath("github.com/tyru/caw.vim")
			out, err := testutil.RunVolt("profile", "rm", "default", reposPath.String())
			// (A, B)
			testutil.SuccessExit(t, out, err)

			lockJSON, err := lockjson.Read()
			if err != nil {
				t.Error("lockjson.Read() returned non-nil error: " + err.Error())
			}
			reposList := getReposList(t, lockJSON, "default")

			// (a)
			if reposList.Contains(reposPath) {
				t.Errorf("expected '%s' is removed from profile '%s', but not removed", reposPath, "default")
			}
			// (b)
			testNotChangedProfileExcept(t, oldLockJSON, lockJSON, "default")
			// (c)
			if lockJSON.Profiles.FindIndexByName("default") == -1 {
				t.Errorf("expected profile '%s' does exist, but does not exist", "default")
			}
		})
	})

	t.Run("Run `volt profile rm <profile> <repos1> <repos2>` (<profile>: exists, <repos1>,<repos2>: exists)", func(t *testing.T) {
		testProfileMatrix(t, func(t *testing.T, strategy string) {
			// =============== setup =============== //

			testutil.SetUpEnv(t)
			defer testutil.CleanUpEnv(t)

			reposPathList := pathutil.ReposPathList{pathutil.ReposPath("github.com/tyru/caw.vim"), pathutil.ReposPath("github.com/tyru/capture.vim")}
			teardown := testutil.SetUpRepos(t, "caw-and-capture", lockjson.ReposGitType, reposPathList, config.SymlinkBuilder)
			defer teardown()
			testutil.InstallConfig(t, "strategy-"+strategy+".toml")

			oldLockJSON, err := lockjson.Read()
			if err != nil {
				t.Error("lockjson.Read() returned non-nil error: " + err.Error())
			}

			// =============== run =============== //

			args := append([]string{"profile", "rm", "default"}, reposPathList.Strings()...)
			out, err := testutil.RunVolt(args...)
			// (A, B)
			testutil.SuccessExit(t, out, err)

			lockJSON, err := lockjson.Read()
			if err != nil {
				t.Error("lockjson.Read() returned non-nil error: " + err.Error())
			}
			reposList := getReposList(t, lockJSON, "default")

			// (a)
			for _, reposPath := range reposPathList {
				if reposList.Contains(reposPath) {
					t.Errorf("expected '%s' is removed from profile '%s', but not removed", reposPath.String(), "default")
				}
			}
			// (b)
			testNotChangedProfileExcept(t, oldLockJSON, lockJSON, "default")
			// (c)
			if lockJSON.Profiles.FindIndexByName("default") == -1 {
				t.Errorf("expected profile '%s' does exist, but does not exist", "default")
			}
		})
	})

	t.Run("Run `volt profile rm -current <repos>` (<repos>: exists)", func(t *testing.T) {
		testProfileMatrix(t, func(t *testing.T, strategy string) {
			// =============== setup =============== //

			testutil.SetUpEnv(t)
			defer testutil.CleanUpEnv(t)

			reposPathList := pathutil.ReposPathList{pathutil.ReposPath("github.com/tyru/caw.vim")}
			teardown := testutil.SetUpRepos(t, "caw.vim", lockjson.ReposGitType, reposPathList, config.SymlinkBuilder)
			defer teardown()
			testutil.InstallConfig(t, "strategy-"+strategy+".toml")

			oldLockJSON, err := lockjson.Read()
			if err != nil {
				t.Error("lockjson.Read() returned non-nil error: " + err.Error())
			}

			// =============== run =============== //

			reposPath := pathutil.ReposPath("github.com/tyru/caw.vim")
			out, err := testutil.RunVolt("profile", "rm", "-current", reposPath.String())
			// (A, B)
			testutil.SuccessExit(t, out, err)

			lockJSON, err := lockjson.Read()
			if err != nil {
				t.Error("lockjson.Read() returned non-nil error: " + err.Error())
			}
			reposList := getReposList(t, lockJSON, lockJSON.CurrentProfileName)

			// (a)
			if reposList.Contains(reposPath) {
				t.Errorf("expected '%s' is removed from profile '%s', but not removed", reposPath, lockJSON.CurrentProfileName)
			}
			// (b)
			testNotChangedProfileExcept(t, oldLockJSON, lockJSON, lockJSON.CurrentProfileName)
			// (c)
			if lockJSON.Profiles.FindIndexByName(lockJSON.CurrentProfileName) == -1 {
				t.Errorf("expected profile '%s' does exist, but does not exist", lockJSON.CurrentProfileName)
			}
		})
	})

	t.Run("Run `volt profile rm <profile> <repos>` (<profile>: not exist, <repos>: exists)", func(t *testing.T) {
		testProfileMatrix(t, func(t *testing.T, strategy string) {
			// =============== setup =============== //

			testutil.SetUpEnv(t)
			defer testutil.CleanUpEnv(t)

			reposPathList := pathutil.ReposPathList{pathutil.ReposPath("github.com/tyru/caw.vim")}
			teardown := testutil.SetUpRepos(t, "caw.vim", lockjson.ReposGitType, reposPathList, config.SymlinkBuilder)
			defer teardown()
			testutil.InstallConfig(t, "strategy-"+strategy+".toml")

			oldLockJSON, err := lockjson.Read()
			if err != nil {
				t.Error("lockjson.Read() returned non-nil error: " + err.Error())
			}

			// =============== run =============== //

			reposPath := "github.com/tyru/caw.vim"
			out, err := testutil.RunVolt("profile", "rm", "not_existing_profile", reposPath)
			// (!A, !B)
			testutil.FailExit(t, out, err)

			lockJSON, err := lockjson.Read()
			if err != nil {
				t.Error("lockjson.Read() returned non-nil error: " + err.Error())
			}

			// (b)
			testNotChangedProfileExcept(t, oldLockJSON, lockJSON, "not_existing_profile")
			// (!c)
			if lockJSON.Profiles.FindIndexByName("not_existing_profile") != -1 {
				t.Errorf("expected profile '%s' does not exist, but does exist", "not_existing_profile")
			}
		})
	})

	t.Run("Run `volt profile rm <profile> <repos>` (<profile>: exists, <repos>: not exist)", func(t *testing.T) {
		testProfileMatrix(t, func(t *testing.T, strategy string) {
			// =============== setup =============== //

			testutil.SetUpEnv(t)
			defer testutil.CleanUpEnv(t)
			testutil.InstallConfig(t, "strategy-"+strategy+".toml")

			oldLockJSON, err := lockjson.Read()
			if err != nil {
				t.Error("lockjson.Read() returned non-nil error: " + err.Error())
			}

			// =============== run =============== //

			reposPath := pathutil.ReposPath("github.com/tyru/caw.vim")
			out, err := testutil.RunVolt("profile", "rm", "default", reposPath.String())
			// (!A, !B)
			testutil.FailExit(t, out, err)

			lockJSON, err := lockjson.Read()
			if err != nil {
				t.Error("lockjson.Read() returned non-nil error: " + err.Error())
			}
			reposList := getReposList(t, lockJSON, "default")

			// (a)
			if reposList.Contains(reposPath) {
				t.Errorf("expected '%s' does not exist on profile '%s', but appears", reposPath, "default")
			}
			// (b)
			testNotChangedProfileExcept(t, oldLockJSON, lockJSON, "default")
			// (c)
			if lockJSON.Profiles.FindIndexByName("default") == -1 {
				t.Errorf("expected profile '%s' does exist, but does not exist", "default")
			}
		})
	})

	t.Run("Run `volt profile rm <profile> <repos>` (<profile>: not exist, <repos>: not exist)", func(t *testing.T) {
		testProfileMatrix(t, func(t *testing.T, strategy string) {
			// =============== setup =============== //

			testutil.SetUpEnv(t)
			defer testutil.CleanUpEnv(t)
			testutil.InstallConfig(t, "strategy-"+strategy+".toml")

			oldLockJSON, err := lockjson.Read()
			if err != nil {
				t.Error("lockjson.Read() returned non-nil error: " + err.Error())
			}

			// =============== run =============== //

			reposPath := "github.com/tyru/caw.vim"
			out, err := testutil.RunVolt("profile", "rm", "not_existing_profile", reposPath)
			// (!A, !B)
			testutil.FailExit(t, out, err)

			lockJSON, err := lockjson.Read()
			if err != nil {
				t.Error("lockjson.Read() returned non-nil error: " + err.Error())
			}

			// (b)
			testNotChangedProfileExcept(t, oldLockJSON, lockJSON, "not_existing_profile")
			// (!c)
			if lockJSON.Profiles.FindIndexByName("not_existing_profile") != -1 {
				t.Errorf("expected profile '%s' does not exist, but does exist", "default")
			}
		})
	})
}

// ============================================

func getReposList(t *testing.T, lockJSON *lockjson.LockJSON, profileName string) lockjson.ReposList {
	currentProfile, err := lockJSON.Profiles.FindByName(profileName)
	if err != nil {
		t.Error("lockJSON.Profiles.FindByName() returned non-nil error: " + err.Error())
	}
	reposList, err := lockJSON.GetReposListByProfile(currentProfile)
	if err != nil {
		t.Error("lockJSON.GetReposListByProfile() returned non-nil error: " + err.Error())
	}
	return reposList
}

// Fails if profile struct was changed (oldLockJSON and newLockJSON differ).
// But the profile named profileName was ignored.
func testNotChangedProfileExcept(t *testing.T, oldLockJSON *lockjson.LockJSON, newLockJSON *lockjson.LockJSON, profileName string) {
	t.Helper()

	if len(oldLockJSON.Profiles) != len(newLockJSON.Profiles) {
		t.Errorf("expected same profiles number but got: old=%d, new=%d", len(oldLockJSON.Profiles), len(newLockJSON.Profiles))
	}

	for i := range newLockJSON.Profiles {
		newProfile := &newLockJSON.Profiles[i]
		if newProfile.Name == profileName {
			continue
		}

		var newStr string
		if b, err := json.Marshal(newProfile); err != nil {
			t.Error("json.Marshal() returned non-nil error: " + err.Error())
		} else {
			newStr = string(b)
		}

		oldProfile, err := oldLockJSON.Profiles.FindByName(newProfile.Name)
		if err != nil {
			t.Error("oldLockJSON.Profiles.FindByName() returned non-nil error: " + err.Error())
		}
		var oldStr string
		if b, err := json.Marshal(oldProfile); err != nil {
			t.Error("json.Marshal() returned non-nil error: " + err.Error())
		} else {
			oldStr = string(b)
		}

		if oldStr != newStr {
			t.Errorf("expected old/new profiles are same but got:\n  old = %s\n  new = %s", oldStr, newStr)
		}
	}
}

func testProfileMatrix(t *testing.T, f func(*testing.T, string)) {
	for _, strategy := range testutil.AvailableStrategies() {
		t.Run(fmt.Sprintf("strategy=%v", strategy), func(t *testing.T) {
			f(t, strategy)
		})
	}
}
