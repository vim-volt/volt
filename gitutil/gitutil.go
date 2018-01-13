package gitutil

import (
	"errors"
	"regexp"

	"github.com/vim-volt/volt/pathutil"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

var refHeadsRx = regexp.MustCompile(`^refs/heads/(.+)$`)

// If the repository is bare:
//   Return the reference of refs/remotes/origin/{branch}
//   where {branch} is default branch
// If the repository is non-bare:
//   Return the reference of current branch's HEAD
func GetHEAD(reposPath pathutil.ReposPath) (string, error) {
	repos, err := git.PlainOpen(pathutil.FullReposPath(reposPath))
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
	defaultBranch := refHeadsRx.FindStringSubmatch(head.Name().String())
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

func SetUpstreamBranch(r *git.Repository) error {
	cfg, err := r.Config()
	if err != nil {
		return err
	}

	head, err := r.Head()
	if err != nil {
		return err
	}

	refBranch := head.Name().String()
	branch := refHeadsRx.FindStringSubmatch(refBranch)
	if len(branch) == 0 {
		return errors.New("HEAD is not matched to refs/heads/...: " + refBranch)
	}

	sec := cfg.Raw.Section("branch")
	subsec := sec.Subsection(branch[1])
	subsec.AddOption("remote", "origin")
	subsec.AddOption("merge", refBranch)

	return r.Storer.SetConfig(cfg)
}
