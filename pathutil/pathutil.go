package pathutil

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// user/name -> github.com/user/name
// github.com/user/name -> github.com/user/name
// https://github.com/user/name -> github.com/user/name
func NormalizeRepository(rawReposPath string) (string, error) {
	paths := strings.Split(rawReposPath, "/")
	if paths[0] == "github.com" && len(paths) == 3 {
		return rawReposPath, nil
	}
	if len(paths) == 2 {
		return "github.com/" + rawReposPath, nil
	}
	if paths[0] == "https:" || paths[0] == "http:" {
		return strings.Join(paths[len(paths)-3:], "/"), nil
	}
	return "", errors.New("invalid format of repository: " + rawReposPath)
}

func HomeDir() string {
	home := os.Getenv("HOME")
	if home != "" {
		return home
	}

	home = os.Getenv("APPDATA") // windows
	if home != "" {
		return home
	}

	panic("Couldn't look up HOME")
}

func VoltPath() string {
	path := os.Getenv("VOLTPATH")
	if path != "" {
		return path
	}
	return filepath.Join(HomeDir(), "volt")
}

func FullReposPathOf(repos string) string {
	return filepath.Join(VoltPath(), "repos", repos)
}

func CloneURLOf(repos string) string {
	return "https://" + repos
}

func SystemPlugConfOf(filename string) string {
	return filepath.Join(VoltPath(), "plugconf", "system", filename)
}

func RCFileOf(filename string) string {
	return filepath.Join(VoltPath(), "rc", filename)
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

func VimDir() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(HomeDir(), "vimfiles")
	} else {
		return filepath.Join(HomeDir(), ".vim")
	}
}
