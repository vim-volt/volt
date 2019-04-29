package subcmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/vim-volt/volt/internal/testutil"
	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/pathutil"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// Checks:
// (A) Does not show `[ERROR]`, `[WARN]` messages
// (B) Exit with zero status
// (C) Repositories are cloned at `$VOLTPATH/repos/<repos>/`
// (D) Plugconf files are installed at `$VOLTPATH/plugconf/<repos>.vim`
// (E) Directories are copied to `~/.vim/pack/volt/<repos>/`, and the contents are same
// (F) Entries are added to lock.json
// (G) tags files are created at `~/.vim/pack/volt/<repos>/doc/tags`
// (H) Output contains "! {repos} > install failed"
// (I) Output contains "! {repos} > upgrade failed"
// (J) Output contains "# {repos} > no change"
// (K) Output contains "# {repos} > already exists"
// (L) Output contains "+ {repos} > added repository to current profile"
// (M) Output contains "+ {repos} > installed"
// (N) Output contains "* {repos} > updated lock.json revision ({from}..{to})"
// (O) Output contains "* {repos} > upgraded ({from}..{to})"
// (P) Output contains "{repos}: HEAD and locked revision are different ..."

// TODO: Add test cases
// * Specify plugins which have dependency plugins without help (A, B, C, D, E, F, !G) / with help (A, B, C, D, E, F, G)
// * Specify plugins which have dependency plugins and plugins which have no dependency plugins without help (A, B, C, D, E, F, !G) / with help (A, B, C, D, E, F, G)

// Specify one plugin with help (A, B, C, D, E, F, G, M) / without help (A, B, C, D, E, F, !G, M)
func TestVoltGetOnePlugin(t *testing.T) {
	for _, tt := range []struct {
		withHelp  bool
		reposPath pathutil.ReposPath
	}{
		{true, pathutil.ReposPath("github.com/tyru/caw.vim")},
		{false, pathutil.ReposPath("github.com/tyru/dummy")},
	} {
		t.Run(fmt.Sprintf("with help=%v", tt.withHelp), func(t *testing.T) {
			testGetMatrix(t, func(t *testing.T, strategy string) {
				// =============== setup =============== //

				testutil.SetUpEnv(t)
				defer testutil.CleanUpEnv(t)
				testutil.InstallConfig(t, "strategy-"+strategy+".toml")

				// =============== run =============== //

				out, err := testutil.RunVolt("get", tt.reposPath.String())
				// (A, B)
				testutil.SuccessExit(t, out, err)

				// (C)
				reposDir := tt.reposPath.FullPath()
				if !pathutil.Exists(reposDir) {
					t.Error("repos does not exist: " + reposDir)
				}
				_, err = git.PlainOpen(reposDir)
				if err != nil {
					t.Error("not git repository: " + reposDir)
				}

				// (D)
				plugconf := tt.reposPath.Plugconf()
				if !pathutil.Exists(plugconf) {
					t.Error("plugconf does not exist: " + plugconf)
				}
				// TODO: check plugconf has s:config(), s:loaded_on(), depends()

				// (E)
				vimReposDir := tt.reposPath.EncodeToPlugDirName()
				if !pathutil.Exists(vimReposDir) {
					t.Error("vim repos does not exist: " + vimReposDir)
				}

				// (F)
				testReposPathWereAdded(t, tt.reposPath)

				tags := filepath.Join(vimReposDir, "doc", "tags")
				if tt.withHelp {
					// (G)
					if !pathutil.Exists(tags) {
						t.Error("doc/tags was not created: " + tags)
					}
				} else {
					// (!G)
					if pathutil.Exists(tags) {
						t.Error("doc/tags was created: " + tags)
					}
				}
			})
		})
	}
}

// (J, K, L, M, N, O, P)
func TestVoltGetMsg(t *testing.T) {
	testGetMatrix(t, func(t *testing.T, strategy string) {
		// =============== setup =============== //

		testutil.SetUpEnv(t)
		defer testutil.CleanUpEnv(t)
		testutil.InstallConfig(t, "strategy-"+strategy+".toml")
		reposPath := pathutil.ReposPath("github.com/tyru/caw.vim")

		// ===================================
		// Install plugin (installed)
		// ===================================

		out, err := testutil.RunVolt("get", reposPath.String())
		// (A, B)
		testutil.SuccessExit(t, out, err)

		// (M)
		msg := fmt.Sprintf(fmtInstalled, reposPath)
		if !bytes.Contains(out, []byte(msg)) {
			t.Errorf("Output does not contain %q\n%s", msg, string(out))
		}

		// ===================================
		// Install again (already exists)
		// ===================================

		out, err = testutil.RunVolt("get", reposPath.String())
		// (A, B)
		testutil.SuccessExit(t, out, err)

		// (K)
		msg = fmt.Sprintf(fmtAlreadyExists, reposPath)
		if !bytes.Contains(out, []byte(msg)) {
			t.Errorf("Output does not contain %q\n%s", msg, string(out))
		}

		// ===================================
		// Upgrade one plugin (no change)
		// ===================================

		out, err = testutil.RunVolt("get", "-u", reposPath.String())
		// (A, B)
		testutil.SuccessExit(t, out, err)

		// (J)
		msg = fmt.Sprintf(fmtNoChange, reposPath)
		if !bytes.Contains(out, []byte(msg)) {
			t.Errorf("Output does not contain %q\n%s", msg, string(out))
		}

		// ===================================
		// Remove plugin from current profile
		// ===================================

		out, err = testutil.RunVolt("disable", reposPath.String())
		// (A, B)
		testutil.SuccessExit(t, out, err)

		// =====================================================================
		// Add plugin to current profile (added repository to current profile")
		// =====================================================================

		out, err = testutil.RunVolt("get", reposPath.String())
		// (A, B)
		testutil.SuccessExit(t, out, err)

		// (L)
		msg = fmt.Sprintf(fmtAddedRepos, reposPath)
		if !bytes.Contains(out, []byte(msg)) {
			t.Errorf("Output does not contain %q\n%s", msg, string(out))
		}

		// ================
		// Commit on repos
		// ================

		head, next, err := gitCommitOne(reposPath)
		if err != nil {
			t.Error("gitCommitOne() failed: " + err.Error())
		}

		// ================================================================
		// volt build outputs "HEAD and locked revision are different ..."
		// ================================================================

		out, err = testutil.RunVolt("build")
		outstr := string(out)
		// (!A, B)
		if !strings.Contains(outstr, "[WARN]") && !strings.Contains(outstr, "[ERROR]") {
			t.Errorf("expected error but no error: %s", outstr)
		}
		if err != nil {
			t.Error("expected success exit but exited with failure: " + err.Error())
		}

		// (P)
		for _, msg := range []string{
			string(reposPath) + ": HEAD and locked revision are different",
			"  HEAD: " + next.String(),
			"  locked revision: " + head.String(),
			"  Please run 'volt get -l' to update locked revision.",
		} {
			if !bytes.Contains(out, []byte(msg)) {
				t.Errorf("Output does not contain %q\n%s", msg, string(out))
			}
		}

		// ===========================================
		// Update lock.json revision
		// ===========================================

		out, err = testutil.RunVolt("get", reposPath.String())
		// (A, B)
		testutil.SuccessExit(t, out, err)

		// (N)
		msg = fmt.Sprintf(fmtRevUpdate, reposPath, head.String(), next.String())
		if !bytes.Contains(out, []byte(msg)) {
			t.Errorf("Output does not contain %q\n%s", msg, string(out))
		}

		// ================================
		// Install again (already exists)
		// ================================

		out, err = testutil.RunVolt("get", reposPath.String())
		// (A, B)
		testutil.SuccessExit(t, out, err)

		// (K)
		msg = fmt.Sprintf(fmtAlreadyExists, reposPath)
		if !bytes.Contains(out, []byte(msg)) {
			t.Errorf("Output does not contain %q\n%s", msg, string(out))
		}

		// ========================================================================
		// volt build DOES NOT output "HEAD and locked revision are different ..."
		// ========================================================================

		out, err = testutil.RunVolt("build")
		// (A, B)
		testutil.SuccessExit(t, out, err)

		// (!P)
		msg = "HEAD and locked revision are different"
		if bytes.Contains(out, []byte(msg)) {
			t.Errorf("Output contains %q\n%s", msg, string(out))
		}

		// ==================================
		// "git reset --hard HEAD~2" on repos
		// ==================================

		prev, _, err := gitResetHard(reposPath, "HEAD~2")
		if err != nil {
			t.Error("gitResetHard() failed: " + err.Error())
		}

		// ===========================================
		// Update lock.json revision
		// ===========================================

		out, err = testutil.RunVolt("get", reposPath.String())
		// (A, B)
		testutil.SuccessExit(t, out, err)

		// (N)
		msg = fmt.Sprintf(fmtRevUpdate, reposPath, next.String(), prev.String())
		if !bytes.Contains(out, []byte(msg)) {
			t.Errorf("Output does not contain %q\n%s", msg, string(out))
		}

		// ================================
		// Upgrade plugin (upgraded)
		// ================================

		out, err = testutil.RunVolt("get", "-u", reposPath.String())
		// (A, B)
		testutil.SuccessExit(t, out, err)

		// (O)
		msg = fmt.Sprintf(fmtUpgraded, reposPath, prev.String(), head.String())
		if !bytes.Contains(out, []byte(msg)) {
			t.Errorf("Output does not contain %q\n%s", msg, string(out))
		}
	})
}

// Specify two or more plugins without help (A, B, C, D, E, F, !G, M) / with help (A, B, C, D, E, F, G, M)
func TestVoltGetTwoOrMorePlugin(t *testing.T) {
	for _, tt := range []struct {
		withHelp      bool
		reposPathList pathutil.ReposPathList
	}{
		{true, pathutil.ReposPathList{
			pathutil.ReposPath("github.com/tyru/caw.vim"),
			pathutil.ReposPath("github.com/tyru/capture.vim"),
		}},
		{false, pathutil.ReposPathList{
			pathutil.ReposPath("github.com/tyru/dummy"),
			pathutil.ReposPath("github.com/tyru/dummy2"),
		}},
	} {
		t.Run(fmt.Sprintf("with help=%v", tt.withHelp), func(t *testing.T) {
			testGetMatrix(t, func(t *testing.T, strategy string) {
				// =============== setup =============== //

				testutil.SetUpEnv(t)
				defer testutil.CleanUpEnv(t)
				testutil.InstallConfig(t, "strategy-"+strategy+".toml")

				// =============== run =============== //

				// (A, B)
				args := append([]string{"get"}, tt.reposPathList.Strings()...)
				out, err := testutil.RunVolt(args...)
				testutil.SuccessExit(t, out, err)

				for _, reposPath := range tt.reposPathList {
					// (C)
					reposDir := reposPath.FullPath()
					if !pathutil.Exists(reposDir) {
						t.Error("repos does not exist: " + reposDir)
					}
					_, err := git.PlainOpen(reposDir)
					if err != nil {
						t.Error("not git repository: " + reposDir)
					}

					// (D)
					plugconf := reposPath.Plugconf()
					if !pathutil.Exists(plugconf) {
						t.Error("plugconf does not exist: " + plugconf)
					}
					// TODO: check plugconf has s:config(), s:loaded_on(), depends()

					// (E)
					vimReposDir := reposPath.EncodeToPlugDirName()
					if !pathutil.Exists(vimReposDir) {
						t.Error("vim repos does not exist: " + vimReposDir)
					}

					// (F)
					testReposPathWereAdded(t, reposPath)

					// (G) and (!G)
					tags := filepath.Join(vimReposDir, "doc", "tags")
					if tt.withHelp {
						if !pathutil.Exists(tags) {
							t.Error("doc/tags was not created: " + tags)
						}
					} else {
						if pathutil.Exists(tags) {
							t.Error("doc/tags was created: " + tags)
						}
					}

					// (M)
					msg := fmt.Sprintf(fmtInstalled, reposPath)
					if !bytes.Contains(out, []byte(msg)) {
						t.Errorf("Output does not contain %q\n%s", msg, string(out))
					}
				}
			})
		})
	}
}

// 'volt get -l' must not add disabled plugins
func TestVoltGetLoptMustNotAddDisabledPlugins(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)
	defer testutil.CleanUpEnv(t)

	// =============== run =============== //

	// (A, B)
	out, err := testutil.RunVolt("get", "tyru/dummy", "tyru/dummy2")
	testutil.SuccessExit(t, out, err)

	testReposPathWereEnabled(t, pathutil.ReposPath("github.com/tyru/dummy"))
	testReposPathWereEnabled(t, pathutil.ReposPath("github.com/tyru/dummy2"))

	// (A, B)
	out, err = testutil.RunVolt("disable", "tyru/dummy2")
	testutil.SuccessExit(t, out, err)

	testReposPathWereEnabled(t, pathutil.ReposPath("github.com/tyru/dummy"))
	testReposPathWereDisabled(t, pathutil.ReposPath("github.com/tyru/dummy2"))

	// (A, B)
	out, err = testutil.RunVolt("get", "-l")
	testutil.SuccessExit(t, out, err)

	testReposPathWereEnabled(t, pathutil.ReposPath("github.com/tyru/dummy"))
	testReposPathWereDisabled(t, pathutil.ReposPath("github.com/tyru/dummy2"))
}

// [error] Specify invalid argument (!A, !B, !C, !D, !E, !F, !G)
func TestErrVoltGetInvalidArgs(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)
	defer testutil.CleanUpEnv(t)

	// =============== run =============== //

	out, err := testutil.RunVolt("get", "caw.vim")
	// (!A, !B)
	testutil.FailExit(t, out, err)

	for _, reposPath := range []pathutil.ReposPath{
		pathutil.ReposPath("caw.vim"),
		pathutil.ReposPath("github.com/caw.vim"),
	} {
		// (!C)
		reposDir := reposPath.FullPath()
		if pathutil.Exists(reposDir) {
			t.Error("repos exists: " + reposDir)
		}

		// (!D)
		plugconf := reposPath.Plugconf()
		if pathutil.Exists(plugconf) {
			t.Error("plugconf exists: " + plugconf)
		}

		// (!E)
		vimReposDir := reposPath.EncodeToPlugDirName()
		if pathutil.Exists(vimReposDir) {
			t.Error("vim repos exists: " + vimReposDir)
		}

		// (!F)
		testReposPathWereNotAdded(t, reposPath)

		// (!G)
		tags := filepath.Join(vimReposDir, "doc", "tags")
		if pathutil.Exists(tags) {
			t.Error("doc/tags was created: " + tags)
		}
	}
}

// [error] Specify plugin which does not exist (!A, !B, !C, !D, !E, !F, !G, H)
func TestErrVoltGetNotFound(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)
	defer testutil.CleanUpEnv(t)

	// =============== run =============== //

	out, err := testutil.RunVolt("get", "vim-volt/not_found")
	// (!A, !B)
	testutil.FailExit(t, out, err)
	reposPath := pathutil.ReposPath("github.com/vim-volt/not_found")

	// (!C)
	reposDir := reposPath.FullPath()
	if pathutil.Exists(reposDir) {
		t.Error("repos exists: " + reposDir)
	}

	// (!D)
	plugconf := reposPath.Plugconf()
	if pathutil.Exists(plugconf) {
		t.Error("plugconf exists: " + plugconf)
	}

	// (!E)
	vimReposDir := reposPath.EncodeToPlugDirName()
	if pathutil.Exists(vimReposDir) {
		t.Error("vim repos exists: " + vimReposDir)
	}

	// (!F)
	testReposPathWereNotAdded(t, reposPath)

	// (!G)
	tags := filepath.Join(vimReposDir, "doc", "tags")
	if pathutil.Exists(tags) {
		t.Error("doc/tags was created: " + tags)
	}

	// (H)
	msg := fmt.Sprintf(fmtInstallFailed, reposPath)
	if !bytes.Contains(out, []byte(msg)) {
		t.Errorf("Output does not contain %q\n%s", msg, string(out))
	}
}

func testReposPathWereAdded(t *testing.T, reposPath pathutil.ReposPath) {
	t.Helper()
	lockJSON, err := lockjson.Read()
	if err != nil {
		t.Error("lockjson.Read() returned non-nil error: " + err.Error())
	}
	if !lockJSON.Repos.Contains(reposPath) {
		t.Error("repos was not added to lock.json/repos: " + reposPath)
	}
	for i := range lockJSON.Profiles {
		if !lockJSON.Profiles[i].ReposPath.Contains(reposPath) {
			t.Error("repos was not added to lock.json/profiles/repos_path: " + reposPath)
		}
	}
}

func testReposPathWereEnabled(t *testing.T, reposPath pathutil.ReposPath) {
	t.Helper()
	checkTestReposPath(t, reposPath, true)
}

func testReposPathWereDisabled(t *testing.T, reposPath pathutil.ReposPath) {
	t.Helper()
	checkTestReposPath(t, reposPath, false)
}

func checkTestReposPath(t *testing.T, reposPath pathutil.ReposPath, enabled bool) {
	t.Helper()
	lockJSON, err := lockjson.Read()
	if err != nil {
		t.Error("lockjson.Read() returned non-nil error: " + err.Error())
	}
	if !lockJSON.Repos.Contains(reposPath) {
		t.Error("repos was not added to lock.json/repos: " + reposPath)
	}
	for i := range lockJSON.Profiles {
		if enabled {
			if !lockJSON.Profiles[i].ReposPath.Contains(reposPath) {
				t.Error("repos was not added to lock.json/profiles/repos_path: " + reposPath)
			}
		} else {
			if lockJSON.Profiles[i].ReposPath.Contains(reposPath) {
				t.Error("repos was added to lock.json/profiles/repos_path: " + reposPath)
			}
		}
	}
}

func testReposPathWereNotAdded(t *testing.T, reposPath pathutil.ReposPath) {
	t.Helper()
	lockJSON, err := lockjson.Read()
	if err != nil {
		t.Error("lockjson.Read() returned non-nil error: " + err.Error())
	}
	if lockJSON.Repos.Contains(reposPath) {
		t.Error("repos was added to lock.json/repos: " + reposPath)
	}
	for i := range lockJSON.Profiles {
		if lockJSON.Profiles[i].ReposPath.Contains(reposPath) {
			t.Error("repos was added to lock.json/profiles/repos_path: " + reposPath)
		}
	}
}

func testGetMatrix(t *testing.T, f func(*testing.T, string)) {
	for _, strategy := range testutil.AvailableStrategies() {
		t.Run(fmt.Sprintf("strategy=%v", strategy), func(t *testing.T) {
			f(t, strategy)
		})
	}
}

func gitCommitOne(reposPath pathutil.ReposPath) (prev plumbing.Hash, current plumbing.Hash, err error) {
	var (
		relPath       = "hello"
		content       = []byte("hello world!")
		commitMsg     = "hello world"
		commitOptions = &git.CommitOptions{
			Author: &object.Signature{
				Name:  "John Doe",
				Email: "john@doe.org",
				When:  time.Now(),
			},
		}
	)

	filename := filepath.Join(reposPath.FullPath(), relPath)
	if err = ioutil.WriteFile(filename, content, 0644); err != nil {
		err = errors.Wrap(err, "ioutil.WriteFile() failed")
		return
	}
	r, err := git.PlainOpen(reposPath.FullPath())
	if err != nil {
		return
	}
	head, err := r.Head()
	if err != nil {
		return
	} else {
		prev = head.Hash()
	}
	w, err := r.Worktree()
	if err != nil {
		return
	}
	_, err = w.Add(relPath)
	if err != nil {
		return
	}
	current, err = w.Commit(commitMsg, commitOptions)
	return
}

func gitResetHard(reposPath pathutil.ReposPath, ref string) (current plumbing.Hash, next plumbing.Hash, err error) {
	r, err := git.PlainOpen(reposPath.FullPath())
	if err != nil {
		return
	}
	w, err := r.Worktree()
	if err != nil {
		return
	}
	head, err := r.Head()
	if err != nil {
		return
	} else {
		next = head.Hash()
	}
	rev, err := r.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return
	} else {
		current = *rev
	}
	err = w.Reset(&git.ResetOptions{
		Commit: current,
		Mode:   git.HardReset,
	})
	return
}
