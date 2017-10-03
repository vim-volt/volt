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
// localhost/local/name -> localhost/local/name
func NormalizeRepos(rawReposPath string) (string, error) {
	rawReposPath = filepath.ToSlash(rawReposPath)
	paths := strings.Split(rawReposPath, "/")
	if len(paths) == 3 {
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

func NormalizeLocalRepos(name string) (string, error) {
	if !strings.Contains(name, "/") {
		return "localhost/local/" + name, nil
	} else {
		return NormalizeRepos(name)
	}
}

func HomeDir() string {
	home := os.Getenv("HOME")
	if home != "" {
		return home
	}

	home = os.Getenv("USERPROFILE") // windows
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
	reposList := strings.Split(filepath.ToSlash(repos), "/")
	paths := make([]string, 0, len(reposList)+2)
	paths = append(paths, VoltPath())
	paths = append(paths, "repos")
	paths = append(paths, reposList...)
	return filepath.Join(paths...)
}

func CloneURLOf(repos string) string {
	return "https://" + filepath.ToSlash(repos)
}

func SystemPlugConfOf(filename string) string {
	filenameList := strings.Split(filepath.ToSlash(filename), "/")
	paths := make([]string, 0, len(filenameList)+3)
	paths = append(paths, VoltPath())
	paths = append(paths, "plugconf")
	paths = append(paths, "system")
	paths = append(paths, filenameList...)
	return filepath.Join(paths...)
}

func RCFileOf(profileName, filename string) string {
	filenameList := strings.Split(filepath.ToSlash(filename), "/")
	paths := make([]string, 0, len(filenameList)+3)
	paths = append(paths, VoltPath())
	paths = append(paths, "rc")
	paths = append(paths, profileName)
	paths = append(paths, filenameList...)
	return filepath.Join(paths...)
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

func VimVoltDir() string {
	return filepath.Join(VimDir(), "pack", "volt")
}

func VimVoltStartDir() string {
	return filepath.Join(VimDir(), "pack", "volt", "start")
}

func BuildInfoJSON() string {
	return filepath.Join(VimVoltStartDir(), "build-info.json")
}

func LookUpVimrcOrGvimrc() []string {
	return append(LookUpVimrc(), LookUpGvimrc()...)
}

func LookUpVimrc() []string {
	var vimrcPaths []string
	if runtime.GOOS == "windows" {
		vimrcPaths = []string{
			filepath.Join(HomeDir(), "_vimrc"),
			filepath.Join(HomeDir(), "vimfiles", "vimrc"),
		}
	} else {
		vimrcPaths = []string{
			filepath.Join(HomeDir(), ".vimrc"),
			filepath.Join(HomeDir(), ".vim", "vimrc"),
		}
	}
	rclist := make([]string, 0, len(vimrcPaths))
	for i := range vimrcPaths {
		if Exists(vimrcPaths[i]) {
			rclist = append(rclist, vimrcPaths[i])
		}
	}
	return rclist
}

func LookUpGvimrc() []string {
	var gvimrcPaths []string
	if runtime.GOOS == "windows" {
		gvimrcPaths = []string{
			filepath.Join(HomeDir(), "_gvimrc"),
			filepath.Join(HomeDir(), "vimfiles", "gvimrc"),
		}
	} else {
		gvimrcPaths = []string{
			filepath.Join(HomeDir(), ".gvimrc"),
			filepath.Join(HomeDir(), ".vim", "gvimrc"),
		}
	}
	rclist := make([]string, 0, len(gvimrcPaths))
	for i := range gvimrcPaths {
		if Exists(gvimrcPaths[i]) {
			rclist = append(rclist, gvimrcPaths[i])
		}
	}
	return rclist
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
