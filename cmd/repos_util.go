package cmd

import (
	"errors"
	"regexp"

	"github.com/vim-volt/volt/pathutil"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

var defaultBranchRx = regexp.MustCompile(`^refs/heads/(.+)$`)

// If the repository is bare:
//   Return the reference of refs/remotes/origin/{branch}
//   where {branch} is default branch
// If the repository is non-bare:
//   Return the reference of current branch's HEAD
func getReposHEAD(reposPath string) (string, error) {
	repos, err := git.PlainOpen(pathutil.FullReposPathOf(reposPath))
	if err != nil {
		return "", err
	}

	head, err := repos.Head()
	if err != nil {
		return "", err
	}

	cfg, err := repos.Config()
	if err != nil {
		return "", err
	}

	if !cfg.Core.IsBare {
		// Get reference of local {branch} HEAD
		commit, err := repos.CommitObject(head.Hash())
		if err != nil {
			return "", err
		}
		return commit.Hash.String(), nil
	}

	// Get branch name from head.Name().String()
	// e.g. head.Name().String() = "refs/heads/master"
	defaultBranch := defaultBranchRx.FindStringSubmatch(head.Name().String())
	if len(defaultBranch) == 0 {
		return "", errors.New("could not find branch name from HEAD")
	}

	// Get reference of remote origin/{branch} HEAD
	ref, err := repos.Reference(plumbing.ReferenceName("refs/remotes/origin/"+defaultBranch[1]), true)
	if err != nil {
		return "", err
	}
	return ref.Hash().String(), nil
}
