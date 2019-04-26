package gitutil

import (
	"regexp"

	"github.com/pkg/errors"

	"github.com/vim-volt/volt/pathutil"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

var refHeadsRx = regexp.MustCompile(`^refs/heads/(.+)$`)

// GetHEAD gets HEAD reference hash string from reposPath.
// See GetHEADRepository.
func GetHEAD(reposPath pathutil.ReposPath) (string, error) {
	repos, err := git.PlainOpen(reposPath.FullPath())
	if err != nil {
		return "", err
	}
	return GetHEADRepository(repos)
}

// GetHEADRepository gets HEAD reference hash string from git.Repository.
// If the repository is bare:
//   Return the reference of refs/remotes/origin/{branch}
//   where {branch} is default branch
// If the repository is non-bare:
//   Return the reference of current branch's HEAD
func GetHEADRepository(repos *git.Repository) (string, error) {
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

// SetUpstreamRemote sets current branch's upstream remote name to remote.
func SetUpstreamRemote(r *git.Repository, remote string) error {
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

	subsec := cfg.Raw.Section("branch").Subsection(branch[1])
	subsec.AddOption("remote", remote)
	subsec.AddOption("merge", refBranch)

	return r.Storer.SetConfig(cfg)
}

// GetUpstreamRemote gets current branch's upstream remote name (e.g. "origin").
func GetUpstreamRemote(r *git.Repository) (string, error) {
	cfg, err := r.Config()
	if err != nil {
		return "", err
	}

	head, err := r.Head()
	if err != nil {
		return "", err
	}

	refBranch := head.Name().String()
	branch := refHeadsRx.FindStringSubmatch(refBranch)
	if len(branch) == 0 {
		return "", errors.New("HEAD is not matched to refs/heads/...: " + refBranch)
	}

	subsec := cfg.Raw.Section("branch").Subsection(branch[1])
	remote := subsec.Option("remote")
	if remote == "" {
		return "", errors.Errorf("gitconfig 'branch.%s.remote' is not found", subsec.Name)
	}
	return remote, nil
}
