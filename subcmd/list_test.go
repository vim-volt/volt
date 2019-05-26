package subcmd

import (
	"strconv"
	"testing"

	"github.com/vim-volt/volt/internal/testutil"
	"github.com/vim-volt/volt/lockjson"
)

// Checks:
// (A) Does not show `[ERROR]`, `[WARN]` messages
// (B) Exit with zero status

// Checks:
// (a) `volt list` and `volt profile show -current` output is same
func TestVoltListAndVoltProfileAreSame(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)
	defer testutil.CleanUpEnv(t)

	// =============== run =============== //

	out1, err := testutil.RunVolt("list")
	testutil.SuccessExit(t, out1, err) // (A, B)
	out2, err := testutil.RunVolt("profile", "show", "-current")
	testutil.SuccessExit(t, out2, err)

	// (a)
	if string(out1) != string(out2) {
		t.Errorf("=== expected ===\n[%s]\n=== got ===\n[%s]", string(out2), string(out1))
	}
}

// Checks:
// (a) `currentProfile` returns current profile
// (b) `version` returns current version
// (c) `versionMajor` returns current version
// (d) `versionMinor` returns current version
// (e) `versionPatch` returns current version
func TestVoltListFunctions(t *testing.T) {
	t.Run("json", func(t *testing.T) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)

		// =============== run =============== //

		out, err := testutil.RunVolt("list", "-f", "{{ json .CurrentProfileName }}")
		// (A, B)
		testutil.SuccessExit(t, out, err)

		// (a)
		if string(out) != "\"default\"" {
			t.Errorf("expected %q but got %q", "\"default\"", string(out))
		}
	})

	t.Run("currentProfile", func(t *testing.T) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)

		// =============== run =============== //

		out, err := testutil.RunVolt("list", "-f", "{{ currentProfile.Name }}")
		// (A, B)
		testutil.SuccessExit(t, out, err)

		lockJSON, err := lockjson.Read()
		if err != nil {
			t.Fatal("failed to read lock.json: " + err.Error())
		}
		// (a)
		if string(out) != lockJSON.CurrentProfileName {
			t.Errorf("expected %q but got %q", lockJSON.CurrentProfileName, string(out))
		}
	})

	t.Run("profile", func(t *testing.T) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)

		// =============== run =============== //

		out, err := testutil.RunVolt("list", "-f", "{{ with profile \"default\" }}{{ .Name }}{{ end }}")
		// (A, B)
		testutil.SuccessExit(t, out, err)

		// (a)
		if string(out) != "default" {
			t.Errorf("expected %q but got %q", "default", string(out))
		}
	})

	t.Run("version", func(t *testing.T) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)

		// =============== run =============== //

		out, err := testutil.RunVolt("list", "-f", "{{ version }}")
		// (A, B)
		testutil.SuccessExit(t, out, err)

		// (b)
		if string(out) != voltVersion {
			t.Errorf("expected %q but got %q", voltVersion, string(out))
		}
	})

	t.Run("versionMajor", func(t *testing.T) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)

		// =============== run =============== //

		out, err := testutil.RunVolt("list", "-f", "{{ versionMajor }}")
		// (A, B)
		testutil.SuccessExit(t, out, err)

		// (c)
		expected := strconv.Itoa(voltVersionInfo()[0])
		if string(out) != expected {
			t.Errorf("expected %q but got %q", expected, string(out))
		}
	})

	t.Run("versionMinor", func(t *testing.T) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)

		// =============== run =============== //

		out, err := testutil.RunVolt("list", "-f", "{{ versionMinor }}")
		// (A, B)
		testutil.SuccessExit(t, out, err)

		// (d)
		expected := strconv.Itoa(voltVersionInfo()[1])
		if string(out) != expected {
			t.Errorf("expected %q but got %q", expected, string(out))
		}
	})

	t.Run("versionPatch", func(t *testing.T) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)

		// =============== run =============== //

		out, err := testutil.RunVolt("list", "-f", "{{ versionPatch }}")
		// (A, B)
		testutil.SuccessExit(t, out, err)

		// (e)
		expected := strconv.Itoa(voltVersionInfo()[2])
		if string(out) != expected {
			t.Errorf("expected %q but got %q", expected, string(out))
		}
	})
}
