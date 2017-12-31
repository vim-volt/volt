package cmd

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/vim-volt/volt/internal/testutil"
	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/pathutil"
	git "gopkg.in/src-d/go-git.v4"
)

// Checks:
// (A) Does not show `[ERROR]`, `[WARN]` messages
// (B) Exit with zero status
// (C) Repositories are cloned at `$VOLTPATH/repos/<repos>/`
// (D) Plugconf files are installed at `$VOLTPATH/plugconf/<repos>.vim`
// (E) Directories are copied to `~/.vim/pack/volt/<repos>/`, and the contents are same
// (F) Entries are added to lock.json
// (G) tags files are created at `~/.vim/pack/volt/<repos>/doc/tags`

// TODO: Add test cases
// * Specify plugins which have dependency plugins without help (A, B, C, D, E, F, !G) / with help (A, B, C, D, E, F, G)
// * Specify plugins which have dependency plugins and plugins which have no dependency plugins without help (A, B, C, D, E, F, !G) / with help (A, B, C, D, E, F, G)

// Specify one plugin with help (A, B, C, D, E, F, G) / without help (A, B, C, D, E, F, !G)
func TestVoltGetOnePlugin(t *testing.T) {
	for _, tt := range []struct {
		withHelp  bool
		reposPath string
	}{
		{true, "github.com/tyru/caw.vim"},
		{false, "github.com/tyru/dummy"},
	} {
		t.Run(fmt.Sprintf("with help=%v", tt.withHelp), func(t *testing.T) {
			testGetMatrix(t, func(t *testing.T, strategy string) {
				// =============== setup =============== //

				testutil.SetUpEnv(t)
				testutil.InstallConfig(t, "strategy-"+strategy+".toml")

				// =============== run =============== //

				out, err := testutil.RunVolt("get", tt.reposPath)
				// (A, B)
				testutil.SuccessExit(t, out, err)

				// (C)
				reposDir := pathutil.FullReposPathOf(tt.reposPath)
				if !pathutil.Exists(reposDir) {
					t.Error("repos does not exist: " + reposDir)
				}
				_, err = git.PlainOpen(reposDir)
				if err != nil {
					t.Error("not git repository: " + reposDir)
				}

				// (D)
				plugconf := pathutil.PlugconfOf(tt.reposPath)
				if !pathutil.Exists(plugconf) {
					t.Error("plugconf does not exist: " + plugconf)
				}
				// TODO: check plugconf has s:config(), s:loaded_on(), depends()

				// (E)
				vimReposDir := pathutil.PackReposPathOf(tt.reposPath)
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

// Specify two or more plugins without help (A, B, C, D, E, F, !G) / with help (A, B, C, D, E, F, G)
func TestVoltGetTwoOrMorePlugin(t *testing.T) {
	for _, tt := range []struct {
		withHelp      bool
		reposPathList []string
	}{
		{true, []string{"github.com/tyru/caw.vim", "github.com/tyru/capture.vim"}},
		{false, []string{"github.com/tyru/dummy", "github.com/tyru/dummy2"}},
	} {
		t.Run(fmt.Sprintf("with help=%v", tt.withHelp), func(t *testing.T) {
			testGetMatrix(t, func(t *testing.T, strategy string) {
				// =============== setup =============== //

				testutil.SetUpEnv(t)
				testutil.InstallConfig(t, "strategy-"+strategy+".toml")

				// =============== run =============== //

				// (A, B)
				args := append([]string{"get"}, tt.reposPathList...)
				out, err := testutil.RunVolt(args...)
				testutil.SuccessExit(t, out, err)

				for _, reposPath := range tt.reposPathList {
					// (C)
					reposDir := pathutil.FullReposPathOf(reposPath)
					if !pathutil.Exists(reposDir) {
						t.Error("repos does not exist: " + reposDir)
					}
					_, err := git.PlainOpen(reposDir)
					if err != nil {
						t.Error("not git repository: " + reposDir)
					}

					// (D)
					plugconf := pathutil.PlugconfOf(reposPath)
					if !pathutil.Exists(plugconf) {
						t.Error("plugconf does not exist: " + plugconf)
					}
					// TODO: check plugconf has s:config(), s:loaded_on(), depends()

					// (E)
					vimReposDir := pathutil.PackReposPathOf(reposPath)
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
				}
			})
		})
	}
}

// [error] Specify invalid argument (!A, !B, !C, !D, !E, !F, !G)
func TestErrVoltGetInvalidArgs(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)

	// =============== run =============== //

	out, err := testutil.RunVolt("get", "caw.vim")
	// (!A, !B)
	testutil.FailExit(t, out, err)

	for _, reposPath := range []string{"caw.vim", "github.com/caw.vim"} {
		// (!C)
		reposDir := pathutil.FullReposPathOf(reposPath)
		if pathutil.Exists(reposDir) {
			t.Error("repos exists: " + reposDir)
		}

		// (!D)
		plugconf := pathutil.PlugconfOf(reposPath)
		if pathutil.Exists(plugconf) {
			t.Error("plugconf exists: " + plugconf)
		}

		// (!E)
		vimReposDir := pathutil.PackReposPathOf(reposPath)
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

// [error] Specify plugin which does not exist (!A, !B, !C, !D, !E, !F, !G)
func TestErrVoltGetNotFound(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)

	// =============== run =============== //

	out, err := testutil.RunVolt("get", "vim-volt/not_found")
	// (!A, !B)
	testutil.FailExit(t, out, err)
	reposPath := "github.com/vim-volt/not_found"

	// (!C)
	reposDir := pathutil.FullReposPathOf(reposPath)
	if pathutil.Exists(reposDir) {
		t.Error("repos exists: " + reposDir)
	}

	// (!D)
	plugconf := pathutil.PlugconfOf(reposPath)
	if pathutil.Exists(plugconf) {
		t.Error("plugconf exists: " + plugconf)
	}

	// (!E)
	vimReposDir := pathutil.PackReposPathOf(reposPath)
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

func testReposPathWereAdded(t *testing.T, reposPath string) {
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

func testReposPathWereNotAdded(t *testing.T, reposPath string) {
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
