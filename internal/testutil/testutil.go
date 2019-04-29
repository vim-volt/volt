package testutil

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/pkg/errors"

	"github.com/vim-volt/volt/config"
	"github.com/vim-volt/volt/fileutil"
	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/pathutil"
)

var voltCommand string
var testdataDir string

func init() {
	const thisFile = "internal/testutil/testutil.go"
	_, fn, _, _ := runtime.Caller(0)
	dir := strings.TrimSuffix(fn, thisFile)

	if runtime.GOOS == "windows" {
		voltCommand = filepath.Join(dir, "bin", "volt.exe")
	} else {
		voltCommand = filepath.Join(dir, "bin", "volt")
	}

	testdataDir = filepath.Join(dir, "testdata")
	os.RemoveAll(filepath.Join(testdataDir, "voltpath"))
}

func TestdataDir() string {
	return testdataDir
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

func CleanUpEnv(t *testing.T) {
	for _, env := range []string{"VOLTPATH", "HOME"} {
		parent, _ := filepath.Split(os.Getenv(env))
		os.RemoveAll(parent)
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
		t.Errorf("expected no error but has error: %s", outstr)
	}
	if err != nil {
		t.Errorf("expected success exit but exited with failure: status=%q, out=%s", err, outstr)
	}
}

func FailExit(t *testing.T, out []byte, err error) {
	t.Helper()
	outstr := string(out)
	if !strings.Contains(outstr, "[WARN]") && !strings.Contains(outstr, "[ERROR]") {
		t.Errorf("expected error but no error: %s", outstr)
	}
	if err == nil {
		t.Errorf("expected failure exit but exited with success: out=%s", outstr)
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

// Set up $VOLTPATH after "volt get <repos>"
// but the same repository is cloned only at first time
// under testdata/voltpath/{testdataName}/repos/<repos>
func SetUpRepos(t *testing.T, testdataName string, rType lockjson.ReposType, reposPathList []pathutil.ReposPath, strategy string) func() {
	voltpath := os.Getenv("VOLTPATH")
	tmpVoltpath := filepath.Join(testdataDir, "voltpath", testdataName)
	localSrcDir := filepath.Join(testdataDir, "local", testdataName)
	localName := fmt.Sprintf("localhost/local/%s", testdataName)
	buf := make([]byte, 32*1024)

	for _, reposPath := range reposPathList {
		testRepos := filepath.Join(tmpVoltpath, "repos", reposPath.String())
		if !pathutil.Exists(testRepos) {
			switch rType {
			case lockjson.ReposGitType:
				home := os.Getenv("HOME")
				tmpHome, err := ioutil.TempDir("", "volt-test-home-")
				if err != nil {
					t.Fatalf("failed to create temp dir: %s", err)
				}
				if err := os.Setenv("HOME", tmpHome); err != nil {
					t.Fatalf("failed to set VOLTPATH: %s", err)
				}
				defer os.Setenv("HOME", home)
				defer os.RemoveAll(home)
				if err := os.Setenv("VOLTPATH", tmpVoltpath); err != nil {
					t.Fatal("failed to set VOLTPATH")
				}
				defer os.Setenv("VOLTPATH", voltpath)
				out, err := RunVolt("get", reposPath.String())
				SuccessExit(t, out, err)
			case lockjson.ReposStaticType:
				err := os.Setenv("VOLTPATH", tmpVoltpath)
				if err != nil {
					t.Fatalf("failed to set VOLTPATH: %s", err)
				}
				defer os.Setenv("VOLTPATH", voltpath)
				os.MkdirAll(filepath.Dir(testRepos), 0777)
				if err := fileutil.CopyDir(localSrcDir, testRepos, buf, 0777, 0); err != nil {
					t.Fatalf("failed to copy %s to %s: %s", localSrcDir, testRepos, err)
				}
				out, err := RunVolt("get", localName)
				SuccessExit(t, out, err)
			default:
				t.Fatalf("unknown type %q", rType)
			}
		}

		// Copy repository
		repos := filepath.Join(voltpath, "repos", reposPath.String())
		os.MkdirAll(filepath.Dir(repos), 0777)
		if err := fileutil.CopyDir(testRepos, repos, buf, 0777, os.FileMode(0)); err != nil {
			t.Fatalf("failed to copy %s to %s: %s", testRepos, repos, err)
		}

		// Copy lock.json
		testLockjsonPath := filepath.Join(tmpVoltpath, "lock.json")
		lockjsonPath := filepath.Join(voltpath, "lock.json")
		os.MkdirAll(filepath.Dir(lockjsonPath), 0777)
		if err := fileutil.CopyFile(testLockjsonPath, lockjsonPath, buf, 0777); err != nil {
			t.Fatalf("failed to copy %s to %s: %s", testLockjsonPath, lockjsonPath, err)
		}
	}
	if strategy == config.SymlinkBuilder {
		return func() {
			for _, reposPath := range reposPathList {
				dir := filepath.Join(tmpVoltpath, "repos", reposPath.String(), "doc")
				for _, name := range []string{"tags", "tags-ja"} {
					path := filepath.Join(dir, name)
					os.Remove(path)
					if pathutil.Exists(path) {
						t.Fatal("could not remove " + path)
					}
				}
			}
		}
	}
	return func() {}
}

func InstallConfig(t *testing.T, filename string) {
	configFile := filepath.Join(testdataDir, "config", filename)
	voltpath := os.Getenv("VOLTPATH")
	dst := filepath.Join(voltpath, "config.toml")
	os.MkdirAll(filepath.Dir(dst), 0777)
	if err := fileutil.CopyFile(configFile, dst, nil, 0777); err != nil {
		t.Fatalf("failed to copy %s to %s", configFile, dst)
	}
}

func DefaultMatrix(t *testing.T, f func(*testing.T, bool, string)) {
	for _, tt := range []struct {
		full     bool
		strategy string
	}{
		{false, config.SymlinkBuilder},
		{false, config.CopyBuilder},
		{true, config.SymlinkBuilder},
		{true, config.CopyBuilder},
	} {
		t.Run(fmt.Sprintf("full=%v,strategy=%v", tt.full, tt.strategy), func(t *testing.T) {
			f(t, tt.full, tt.strategy)
		})
	}
}

func AvailableStrategies() []string {
	return []string{config.SymlinkBuilder, config.CopyBuilder}
}
