package it

import (
	"testing"

	"github.com/vim-volt/volt/internal/testutils"
	"github.com/vim-volt/volt/pathutil"
	git "gopkg.in/src-d/go-git.v4"
)

func TestVoltGetOnePlugin(t *testing.T) {
	testutils.SetUpVoltpath(t)
	out, err := testutils.RunVolt("get", "tyru/caw.vim")
	testutils.SuccessExit(t, out, err)
	reposPath := "github.com/tyru/caw.vim"

	// Check git repository was cloned
	reposDir := pathutil.FullReposPathOf(reposPath)
	if !pathutil.Exists(reposDir) {
		t.Fatal("repos does not exist: " + reposDir)
	}
	_, err = git.PlainOpen(reposDir)
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

func TestVoltGetTwoOrMorePlugin(t *testing.T) {
	testutils.SetUpVoltpath(t)
	out, err := testutils.RunVolt("get", "tyru/caw.vim", "tyru/skk.vim")
	testutils.SuccessExit(t, out, err)
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
}

func TestErrVoltGetInvalidArgs(t *testing.T) {
	testutils.SetUpVoltpath(t)
	out, err := testutils.RunVolt("get", "caw.vim")
	testutils.FailExit(t, out, err)

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
}

func TestErrVoltGetNotFound(t *testing.T) {
	testutils.SetUpVoltpath(t)
	out, err := testutils.RunVolt("get", "vim-volt/not_found")
	testutils.FailExit(t, out, err)
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
}
