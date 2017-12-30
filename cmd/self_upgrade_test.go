package cmd

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestVoltSelfUpgrade(t *testing.T) {
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

	var code int
	out := captureOutput(t, func() {
		code = Run("self-upgrade", []string{"-check"})
	})

	if code != 0 {
		t.Error("Expected exitcode=0, but got: " + strconv.Itoa(code))
	}
	if !strings.Contains(out, "---") {
		t.Error("Expected release notes, but got: " + out)
	}
}

// 'volt self-upgrade -check' from current version should show the latest release
func testVoltSelfUpgradeCheckFromCurrentVer(t *testing.T) {
	var code int
	out := captureOutput(t, func() {
		code = Run("self-upgrade", []string{"-check"})
	})

	if code != 0 {
		t.Error("Expected exitcode=0, but got: " + strconv.Itoa(code))
	}
	if out != "[INFO] No updates were found.\n" {
		t.Error("Expected no updates found, but got: " + out)
	}
}

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
