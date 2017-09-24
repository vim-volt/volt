package pathutil

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// user/name -> github.com/user/name
// github.com/user/name -> github.com/user/name
// https://github.com/user/name -> github.com/user/name
func NormalizeRepository(repos string) (string, error) {
	paths := strings.Split(repos, "/")
	if paths[0] == "github.com" && len(paths) == 3 {
		return repos, nil
	}
	if len(paths) == 2 {
		return "github.com/" + repos, nil
	}
	if paths[0] == "https:" || paths[0] == "http:" {
		return strings.Join(paths[len(paths)-3:], "/"), nil
	}
	return "", errors.New("invalid format of repository")
}

func VoltPath() string {
	path := os.Getenv("VOLTPATH")
	if path != "" {
		return path
	}
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("APPDATA") // windows
		if home == "" {
			panic("Couldn't look up VOLTPATH")
		}
	}
	return filepath.Join(home, "volt")
}

func FullReposPathOf(repos string) string {
	return filepath.Join(VoltPath(), "repos", repos)
}

func CloneURLOf(repos string) string {
	return "https://" + repos
}

func SystemPlugConfOf(filename string) string {
	return filepath.Join(VoltPath(), "plugconf", filename)
}

func LockJSON() string {
	return filepath.Join(VoltPath(), "lock.json")
}

func TrxLock() string {
	return filepath.Join(VoltPath(), "trx.lock")
}

func TempPath() string {
	return filepath.Join(VoltPath(), "tmp")
}
