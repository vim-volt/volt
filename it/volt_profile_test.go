package it

import (
	"testing"

	"github.com/vim-volt/volt/internal/testutils"
	"github.com/vim-volt/volt/lockjson"
)

// Checks:
// (A) Does not show `[ERROR]`, `[WARN]` messages
// (B) Exit with zero status
// (C) Changes current profile

// * Run `volt profile set <profile>` (`<profile>` is not current profile) (A, B, C)
// * Run `volt profile set <profile>` (`<profile>` is current profile) (!A, !B, !C)
func TestVoltProfileSet(t *testing.T) {
	testutils.SetUpEnv(t)
	newOut, newErr := testutils.RunVolt("profile", "new", "foo")
	testutils.SuccessExit(t, newOut, newErr)

	// Run `volt profile set <profile>` (`<profile>` is not current profile)
	profileName := "foo"
	setOut, setErr := testutils.RunVolt("profile", "set", profileName)
	// (A, B)
	testutils.SuccessExit(t, setOut, setErr)

	// (C)
	lockJSON, err := lockjson.Read()
	if err != nil {
		t.Fatal("lockjson.Read() returned non-nil error: " + err.Error())
	}
	if lockJSON.CurrentProfileName != profileName {
		t.Fatal("expected: %s, got: %s", profileName, lockJSON.CurrentProfileName)
	}

	// Run `volt profile set <profile>` (`<profile>` is current profile)
	out, err := testutils.RunVolt("profile", "set", profileName)
	// (!A, !B)
	testutils.FailExit(t, out, err)

	// (!C)
	lockJSON, err = lockjson.Read()
	if err != nil {
		t.Fatal("lockjson.Read() returned non-nil error: " + err.Error())
	}
	if lockJSON.CurrentProfileName != profileName {
		t.Fatal("expected: %s, got: %s", profileName, lockJSON.CurrentProfileName)
	}
}
