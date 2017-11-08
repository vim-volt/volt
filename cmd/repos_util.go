package cmd

import (
	"errors"
	"regexp"

	"github.com/vim-volt/volt/pathutil"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

var refHeadBranchRx = regexp.MustCompile(`^refs/heads/(.+)$`)

func getRemoteHEAD(reposPath string) (string, error) {
	repos, err := git.PlainOpen(pathutil.FullReposPathOf(reposPath))
	if err != nil {
		return "", err
	}

	head, err := repos.Head()
	if err != nil {
		return "", err
	}

	// e.g. head.Name() = "refs/heads/master"
	match := refHeadBranchRx.FindStringSubmatch(head.Name().String())
	if len(match) == 0 {
		return "", errors.New("could not find branch name from HEAD")
	}

	// Get reference of refs/remotes/origin/{branchName}
	ref, err := repos.Reference(plumbing.ReferenceName("refs/remotes/origin/"+match[1]), true)
	if err != nil {
		return "", err
	}

	return ref.Hash().String(), nil
}
