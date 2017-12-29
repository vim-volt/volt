package cmd

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
)

// 'volt self-upgrade -check' from old version should show the latest release
func TestVoltSelfUpgradeCheckFromOldVer(t *testing.T) {
	t.Skip("this test blocks input infinitely...")

	// =============== setup =============== //

	oldVersion := voltVersion
	voltVersion = "v0.0.1"
	defer func() { voltVersion = oldVersion }()

	outCh, errCh, teardown := setupOutput(t)
	defer teardown()

	// =============== run =============== //

	code := Run("self-upgrade", []string{"-check"})

	if code != 0 {
		t.Fatal("Expected exitcode=0, but got: " + strconv.Itoa(code))
	}

	stdout, stderr := captureStdouterr(t, outCh, errCh)

	if !strings.Contains(stdout, "---") {
		t.Fatal("Expected release notes, but got: " + stdout)
	}
	if stderr != "" {
		t.Fatal("Expected no stderr output, but got: " + stderr)
	}
}

// 'volt self-upgrade -check' from current version should show the latest release
func TestVoltSelfUpgradeCheckFromCurrentVer(t *testing.T) {
	t.Skip("this test blocks input infinitely...")

	// =============== setup =============== //

	outCh, errCh, teardown := setupOutput(t)
	defer teardown()

	// =============== run =============== //

	code := Run("self-upgrade", []string{"-check"})

	stdout, stderr := captureStdouterr(t, outCh, errCh)

	if code != 0 {
		t.Fatal("Expected exitcode=0, but got: " + strconv.Itoa(code))
	}
	if stdout != "[INFO] No updates were found.\n" {
		t.Fatal("Expected no updates found, but got: " + stdout)
	}
	if stderr != "" {
		t.Fatal("Expected no stderr output, but got: " + stderr)
	}
}

func setupOutput(t *testing.T) (chan string, chan string, func()) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	stdoutR, stdoutW, err := os.Pipe()
	os.Stdout = stdoutW
	if err != nil {
		t.Fatal("os.Pipe() failed: " + err.Error())
	}
	stderrR, stderrW, err := os.Pipe()
	os.Stderr = stderrW
	if err != nil {
		t.Fatal("os.Pipe() failed: " + err.Error())
	}
	outCh := make(chan string, 1)
	go func() {
		b, err := ioutil.ReadAll(stdoutR)
		if err != nil {
			t.Fatal("ioutil.ReadAll() failed: " + err.Error())
		}
		outCh <- string(b)
	}()
	errCh := make(chan string, 1)
	go func() {
		b, err := ioutil.ReadAll(stderrR)
		if err != nil {
			t.Fatal("ioutil.ReadAll() failed: " + err.Error())
		}
		errCh <- string(b)
	}()
	return outCh, errCh, func() {
		stdoutW.Close()
		os.Stdout = oldStdout
		stderrW.Close()
		os.Stderr = oldStderr
	}
}

func captureStdouterr(t *testing.T, outCh chan string, errCh chan string) (string, string) {
	return <-outCh, <-errCh
}
