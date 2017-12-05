package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/vim-volt/volt/pathutil"
	git "gopkg.in/src-d/go-git.v4"
)

const voltCommand = "../bin/volt"

func TestVoltGetOnePlugin(t *testing.T) {
	out, cmdErr := runVolt(t, "get", "tyru/caw.vim")
	reposPath := "github.com/tyru/caw.vim"

	// Check git repository was cloned
	reposDir := pathutil.FullReposPathOf(reposPath)
	if !pathutil.Exists(reposDir) {
		t.Fatal("repos does not exist: " + reposDir)
	}
	_, err := git.PlainOpen(reposDir)
	if err != nil {
		t.Fatal("not git repository: " + reposDir)
	}

	// Check plugconf was created
	plugconf := pathutil.PlugconfOf(reposPath)
	if !pathutil.Exists(plugconf) {
		t.Fatal("plugconf does not exist: " + plugconf)
	}
	// TODO: check plugconf has s:config(), s:loaded_on(), depends()

	// Check directory was created under vim dir
	vimReposDir := pathutil.PackReposPathOf(reposPath)
	if !pathutil.Exists(vimReposDir) {
		t.Fatal("vim repos does not exist: " + vimReposDir)
	}

	// Check output and exit code
	outstr := string(out)
	if strings.Contains(outstr, "[WARN]") || strings.Contains(outstr, "[ERROR]") {
		t.Fatal("expected no error but has error: " + outstr)
	}
	if cmdErr != nil {
		t.Fatal("cmdErr != nil: " + cmdErr.Error())
	}
}

func TestVoltGetTwoOrMorePlugin(t *testing.T) {
	out, cmdErr := runVolt(t, "get", "tyru/caw.vim", "tyru/skk.vim")
	cawReposPath := "github.com/tyru/caw.vim"
	skkReposPath := "github.com/tyru/skk.vim"

	for _, reposPath := range []string{cawReposPath, skkReposPath} {
		// Check git repository was cloned
		reposDir := pathutil.FullReposPathOf(reposPath)
		if !pathutil.Exists(reposDir) {
			t.Fatal("repos does not exist: " + reposDir)
		}
		_, err := git.PlainOpen(reposDir)
		if err != nil {
			t.Fatal("not git repository: " + reposDir)
		}

		// Check plugconf was created
		plugconf := pathutil.PlugconfOf(reposPath)
		if !pathutil.Exists(plugconf) {
			t.Fatal("plugconf does not exist: " + plugconf)
		}
		// TODO: check plugconf has s:config(), s:loaded_on(), depends()

		// Check directory was created under vim dir
		vimReposDir := pathutil.PackReposPathOf(reposPath)
		if !pathutil.Exists(vimReposDir) {
			t.Fatal("vim repos does not exist: " + vimReposDir)
		}
	}

	// Check output and exit code
	outstr := string(out)
	if strings.Contains(outstr, "[WARN]") || strings.Contains(outstr, "[ERROR]") {
		t.Fatal("expected no error but has error: " + outstr)
	}
	if cmdErr != nil {
		t.Fatal("cmdErr != nil: " + cmdErr.Error())
	}
}

func TestVoltGetInvalidArgs(t *testing.T) {
	out, cmdErr := runVolt(t, "get", "caw.vim")

	// Check repos was not cloned
	reposDir := pathutil.FullReposPathOf("caw.vim")
	if pathutil.Exists(reposDir) {
		t.Fatal("repos exists: " + reposDir)
	}
	reposDir = pathutil.FullReposPathOf("github.com/caw.vim")
	if pathutil.Exists(reposDir) {
		t.Fatal("repos exists: " + reposDir)
	}

	// Check plugconf was not created
	plugconf := pathutil.PlugconfOf("caw.vim")
	if pathutil.Exists(plugconf) {
		t.Fatal("plugconf exists: " + plugconf)
	}
	plugconf = pathutil.PlugconfOf("github.com/caw.vim")
	if pathutil.Exists(plugconf) {
		t.Fatal("plugconf exists: " + plugconf)
	}

	// Check directory was not created under vim dir
	vimReposDir := pathutil.PackReposPathOf("caw.vim")
	if pathutil.Exists(vimReposDir) {
		t.Fatal("vim repos exists: " + vimReposDir)
	}
	vimReposDir = pathutil.PackReposPathOf("github.com/caw.vim")
	if pathutil.Exists(vimReposDir) {
		t.Fatal("vim repos exists: " + vimReposDir)
	}

	// Check output and exit code
	outstr := string(out)
	if !strings.Contains(outstr, "[WARN]") && !strings.Contains(outstr, "[ERROR]") {
		t.Fatal("expected error but no error: " + outstr)
	}
	if cmdErr == nil {
		t.Fatal("cmdErr == nil: " + cmdErr.Error())
	}
}

func TestVoltGetNotFound(t *testing.T) {
	out, cmdErr := runVolt(t, "get", "vim-volt/not_found")
	reposPath := "github.com/vim-volt/not_found"

	// Check repos was not cloned
	reposDir := pathutil.FullReposPathOf(reposPath)
	if pathutil.Exists(reposDir) {
		t.Fatal("repos exists: " + reposDir)
	}

	// Check plugconf was not created
	plugconf := pathutil.PlugconfOf(reposPath)
	if pathutil.Exists(plugconf) {
		t.Fatal("plugconf exists: " + plugconf)
	}

	// Check directory was created under vim dir
	vimReposDir := pathutil.PackReposPathOf(reposPath)
	if pathutil.Exists(vimReposDir) {
		t.Fatal("vim repos exists: " + vimReposDir)
	}

	// Check output and exit code
	outstr := string(out)
	if !strings.Contains(outstr, "[WARN]") && !strings.Contains(outstr, "[ERROR]") {
		t.Fatal("expected error but no error: " + outstr)
	}
	if cmdErr == nil {
		t.Fatal("cmdErr == nil: " + cmdErr.Error())
	}
}

func runVolt(t *testing.T, args ...string) ([]byte, error) {
	cmd := exec.Command(voltCommand, args...)
	tempDir, err := ioutil.TempDir("/tmp", "volt-test-")
	if err != nil {
		t.Fatal("failed to create temp dir")
	}
	err = os.Setenv("VOLTPATH", tempDir)
	if err != nil {
		t.Fatal("failed to set VOLTPATH")
	}
	cmd.Env = append(os.Environ(), "VOLTPATH="+tempDir)
	return cmd.CombinedOutput()
}
