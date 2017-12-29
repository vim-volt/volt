package testutil

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
)

var voltCommand string

func init() {
	const thisFile = "internal/testutil/testutil.go"
	_, fn, _, _ := runtime.Caller(0)
	dir := strings.TrimSuffix(fn, thisFile)
	if runtime.GOOS == "windows" {
		voltCommand = filepath.Join(dir, "bin", "volt.exe")
	} else {
		voltCommand = filepath.Join(dir, "bin", "volt")
	}
}

func SetUpEnv(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "volt-test-")
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
	t.Helper()
	outstr := string(out)
	if strings.Contains(outstr, "[WARN]") || strings.Contains(outstr, "[ERROR]") {
		t.Fatalf("expected no error but has error: %s", outstr)
	}
	if err != nil {
		t.Fatal("expected success exit but exited with failure: " + err.Error())
	}
}

func FailExit(t *testing.T, out []byte, err error) {
	t.Helper()
	outstr := string(out)
	if !strings.Contains(outstr, "[WARN]") && !strings.Contains(outstr, "[ERROR]") {
		t.Fatalf("expected error but no error: %s", outstr)
	}
	if err == nil {
		t.Fatal("expected failure exit but exited with success")
	}
}

// Return sorted list of command names list
func GetCmdList() ([]string, error) {
	out, err := RunVolt("help")
	if err != nil {
		return nil, err
	}
	outstr := string(out)
	lines := strings.Split(outstr, "\n")
	cmdidx := -1
	for i := range lines {
		if lines[i] == "Command" {
			cmdidx = i + 1
			break
		}
	}
	if cmdidx < 0 {
		return nil, errors.New("not found 'Command' line in 'volt help'")
	}
	dup := make(map[string]bool, 20)
	cmdList := make([]string, 0, 20)
	re := regexp.MustCompile(`^  (\S+)`)
	for i := cmdidx; i < len(lines); i++ {
		if m := re.FindStringSubmatch(lines[i]); len(m) != 0 && !dup[m[1]] {
			cmdList = append(cmdList, m[1])
			dup[m[1]] = true
		}
	}
	sort.Strings(cmdList)
	return cmdList, nil
}
