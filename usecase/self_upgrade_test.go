package usecase

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/vim-volt/volt/config"
	"github.com/vim-volt/volt/gateway"
	"github.com/vim-volt/volt/lockjson"
)

func TestVoltSelfUpgrade(t *testing.T) {
	if os.Getenv("DO_TEST_SELF_UPGRADE") == "" {
		t.Skip("skip tests of self-upgrade due to rate limit of GitHub API")
	}
	// Calling subtests serially to control execution order.
	// These tests rewrite os.Stdout/os.Stderr so running them parallely may cause
	// unexpected behavior.
	t.Run("testVoltSelfUpgradeCheckFromOldVer", testVoltSelfUpgradeCheckFromOldVer)
	t.Run("testVoltSelfUpgradeCheckFromCurrentVer", testVoltSelfUpgradeCheckFromCurrentVer)
}

// 'volt self-upgrade -check' from old version should show the latest release
func testVoltSelfUpgradeCheckFromOldVer(t *testing.T) {
	// =============== setup =============== //

	oldVersion := voltVersion
	voltVersion = "v0.0.1"
	defer func() { voltVersion = oldVersion }()

	// =============== run =============== //

	var err *gateway.Error
	out := captureOutput(t, func() {
		err = runVolt(t, "self-upgrade", "-check")
	})

	if err != nil {
		t.Error("Expected nil error, but got: " + err.Error())
	}
	if !strings.Contains(out, "---") {
		t.Error("Expected release notes, but got: " + out)
	}
}

// 'volt self-upgrade -check' from current version should show the latest release
func testVoltSelfUpgradeCheckFromCurrentVer(t *testing.T) {
	var err *Error
	out := captureOutput(t, func() {
		err = runVolt(t, "self-upgrade", "-check")
	})

	if err != nil {
		t.Error("Expected nil error, but got: " + err.Error())
	}
	if out != "[INFO] No updates were found.\n" {
		t.Error("Expected no updates found, but got: " + out)
	}
}

// TODO use https://github.com/rhysd/go-fakeio
func captureOutput(t *testing.T, f func()) string {
	r, w, err := os.Pipe()
	if err != nil {
		t.Error("os.Pipe() failed: " + err.Error())
	}
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	os.Stdout = w
	os.Stderr = w
	outCh := make(chan string, 1)
	go func() {
		b, err := ioutil.ReadAll(r)
		if err != nil {
			t.Error("ioutil.ReadAll() failed: " + err.Error())
		}
		outCh <- string(b)
	}()

	f()

	w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	return <-outCh
}

func runVolt(t *testing.T, cmd string, args ...string) *Error {
	c := LookUpCmd(cmd)
	if c == nil {
		t.Fatal("unknown command '" + cmd + "'")
	}
	lockJSON, err := lockjson.Read()
	if err != nil {
		t.Fatal(errors.Wrap(err, "failed to read lock.json").Error())
	}
	cfg, err := config.Read()
	if err != nil {
		t.Fatal(errors.Wrap(err, "failed to read config.toml").Error())
	}
	return c.Run(&CmdContext{
		Args:     args,
		LockJSON: lockJSON,
		Config:   cfg,
	})
}
