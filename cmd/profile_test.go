package cmd

import (
	"testing"

	"github.com/vim-volt/volt/internal/testutil"
	"github.com/vim-volt/volt/lockjson"
)

// Checks:
// (A) Does not show `[ERROR]`, `[WARN]` messages
// (B) Exit with zero status
// (C) Changes current profile

// * Run `volt profile set <profile>` (`<profile>` is not current profile) (A, B, C)
// * Run `volt profile set <profile>` (`<profile>` is current profile) (!A, !B, !C)
func TestVoltProfileSet(t *testing.T) {
	testutil.SetUpEnv(t)
	newOut, newErr := testutil.RunVolt("profile", "new", "foo")
	testutil.SuccessExit(t, newOut, newErr)

	// Run `volt profile set <profile>` (`<profile>` is not current profile)
	profileName := "foo"
	setOut, setErr := testutil.RunVolt("profile", "set", profileName)
	// (A, B)
	testutil.SuccessExit(t, setOut, setErr)

	// (C)
	lockJSON, err := lockjson.Read()
	if err != nil {
		t.Fatal("lockjson.Read() returned non-nil error: " + err.Error())
	}
	if lockJSON.CurrentProfileName != profileName {
		t.Fatalf("expected: %s, got: %s", profileName, lockJSON.CurrentProfileName)
	}

	// Run `volt profile set <profile>` (`<profile>` is current profile)
	out, err := testutil.RunVolt("profile", "set", profileName)
	// (!A, !B)
	testutil.FailExit(t, out, err)

	// (!C)
	lockJSON, err = lockjson.Read()
	if err != nil {
		t.Fatal("lockjson.Read() returned non-nil error: " + err.Error())
	}
	if lockJSON.CurrentProfileName != profileName {
		t.Fatalf("expected: %s, got: %s", profileName, lockJSON.CurrentProfileName)
	}
}
