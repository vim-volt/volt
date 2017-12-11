package testutil

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const voltCommand = "../bin/volt"

func SetUpEnv(t *testing.T) {
	tempDir, err := ioutil.TempDir("/tmp", "volt-test-")
	if err != nil {
		t.Fatal("failed to create temp dir")
	}
	for _, env := range []string{"VOLTPATH", "HOME"} {
		value := filepath.Join(tempDir, strings.ToLower(env))
		if os.Mkdir(value, 0755) != nil {
			t.Fatalf("failed to mkdir %s: %s", env, value)
		}
		err = os.Setenv(env, value)
		if err != nil {
			t.Fatalf("failed to set %s", env)
		}
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
		t.Fatalf("expected no error but has error at %s: %s", getCallerMsg(), outstr)
	}
	if err != nil {
		t.Fatal("expected success exit but exited with failure: " + err.Error())
	}
}

func FailExit(t *testing.T, out []byte, err error) {
	outstr := string(out)
	if !strings.Contains(outstr, "[WARN]") && !strings.Contains(outstr, "[ERROR]") {
		t.Fatalf("expected error but no error at %s: %s", getCallerMsg(), outstr)
	}
	if err == nil {
		t.Fatal("expected failure exit but exited with success")
	}
}

func getCallerMsg() string {
	_, fn, line, _ := runtime.Caller(2)
	return fmt.Sprintf("[%s:%d]", fn, line)
}
