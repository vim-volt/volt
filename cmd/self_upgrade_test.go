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
	// =============== setup =============== //

	oldVersion := voltVersion
	voltVersion = "v0.0.1"
	defer func() { voltVersion = oldVersion }()

	outCh, done := setupOutput(t)
	defer close(done)

	// =============== run =============== //

	code := Run("self-upgrade", []string{"-check"})
	if code != 0 {
		t.Fatal("Expected exitcode=0, but got: " + strconv.Itoa(code))
	}

	done <- true
	out := <-outCh
	if !strings.Contains(out, "---") {
		t.Fatal("Expected release notes, but got: " + out)
	}
}

// 'volt self-upgrade -check' from current version should show the latest release
func TestVoltSelfUpgradeCheckFromCurrentVer(t *testing.T) {
	// =============== setup =============== //

	outCh, done := setupOutput(t)
	defer close(done)

	// =============== run =============== //

	code := Run("self-upgrade", []string{"-check"})
	if code != 0 {
		t.Fatal("Expected exitcode=0, but got: " + strconv.Itoa(code))
	}

	done <- true
	out := <-outCh
	if out != "[INFO] No updates were found.\n" {
		t.Fatal("Expected no updates found, but got: " + out)
	}
}

func setupOutput(t *testing.T) (chan string, chan bool) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal("os.Pipe() failed: " + err.Error())
	}
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	os.Stdout = w
	os.Stderr = w
	outCh := make(chan string, 1)
	go func() {
		b, err := ioutil.ReadAll(r)
		if err != nil {
			t.Fatal("ioutil.ReadAll() failed: " + err.Error())
		}
		outCh <- string(b)
	}()
	done := make(chan bool)
	go func() {
		<-done
		w.Close()
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()
	return outCh, done
}
