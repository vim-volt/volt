package it

import (
	"testing"

	"github.com/vim-volt/volt/internal/testutils"
	"github.com/vim-volt/volt/lockjson"
)

func TestVoltProfileSet(t *testing.T) {
	testutils.SetUpVoltpath(t)
	newOut, newErr := testutils.RunVolt("profile", "new", "foo")
	testutils.SuccessExit(t, newOut, newErr)
	setOut, setErr := testutils.RunVolt("profile", "set", "foo")
	testutils.SuccessExit(t, setOut, setErr)

	// Check if it changes current profile
	lockJSON, err := lockjson.Read()
	if err != nil {
		t.Fatal("lockjson.Read() returned non-nil error: " + err.Error())
	}
	if lockJSON.ActiveProfile != "foo" {
		t.Fatal("current profile is not foo: " + lockJSON.ActiveProfile)
	}
}

func TestVoltProfileSetCurrentProfile(t *testing.T) {
	testutils.SetUpVoltpath(t)
	out, err := testutils.RunVolt("profile", "set", "default")
	testutils.FailExit(t, out, err)

	// Check if it changes current profile
	lockJSON, err := lockjson.Read()
	if err != nil {
		t.Fatal("lockjson.Read() returned non-nil error: " + err.Error())
	}
	if lockJSON.ActiveProfile != "default" {
		t.Fatalf("current profile was changed: \"%s\" != \"default\"", lockJSON.ActiveProfile)
	}
}
