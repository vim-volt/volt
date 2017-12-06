package testutils

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

const voltCommand = "../bin/volt"

func SetUpVoltpath(t *testing.T) {
	tempDir, err := ioutil.TempDir("/tmp", "volt-test-")
	if err != nil {
		t.Fatal("failed to create temp dir")
	}
	err = os.Setenv("VOLTPATH", tempDir)
	if err != nil {
		t.Fatal("failed to set VOLTPATH")
	}
}

func RunVolt(args ...string) ([]byte, error) {
	cmd := exec.Command(voltCommand, args...)
	// cmd.Env = append(os.Environ(), "VOLTPATH="+voltpath)
	return cmd.CombinedOutput()
}

func SuccessExit(t *testing.T, out []byte, err error) {
	outstr := string(out)
	if strings.Contains(outstr, "[WARN]") || strings.Contains(outstr, "[ERROR]") {
		t.Fatal("expected no error but has error: " + outstr)
	}
	if err != nil {
		t.Fatal("expected success exit but exited with failure: " + err.Error())
	}
}

func FailExit(t *testing.T, out []byte, err error) {
	outstr := string(out)
	if !strings.Contains(outstr, "[WARN]") && !strings.Contains(outstr, "[ERROR]") {
		t.Fatal("expected error but no error: " + outstr)
	}
	if err == nil {
		t.Fatal("expected failure exit but exited with success")
	}
}
