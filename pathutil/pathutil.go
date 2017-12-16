package pathutil

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Normalize the following forms into "github.com/user/name":
// 1. user/name[.git]
// 2. github.com/user/name[.git]
// 3. [git|http|https]://github.com/user/name[.git]
func NormalizeRepos(rawReposPath string) (string, error) {
	rawReposPath = filepath.ToSlash(rawReposPath)
	paths := strings.Split(rawReposPath, "/")
	if len(paths) == 3 {
		return strings.TrimSuffix(rawReposPath, ".git"), nil
	}
	if len(paths) == 2 {
		return strings.TrimSuffix("github.com/"+rawReposPath, ".git"), nil
	}
	if paths[0] == "https:" || paths[0] == "http:" || paths[0] == "git:" {
		reposPath := strings.Join(paths[len(paths)-3:], "/")
		return strings.TrimSuffix(reposPath, ".git"), nil
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

func PlugconfOf(reposPath string) string {
	filenameList := strings.Split(filepath.ToSlash(reposPath+".vim"), "/")
	paths := make([]string, 0, len(filenameList)+2)
	paths = append(paths, VoltPath())
	paths = append(paths, "plugconf")
	paths = append(paths, filenameList...)
	return filepath.Join(paths...)
}

const ProfileVimrc = "vimrc.vim"
const ProfileGvimrc = "gvimrc.vim"
const Vimrc = "vimrc"
const Gvimrc = "gvimrc"

func RCDir(profileName string) string {
	return filepath.Join([]string{VoltPath(), "rc", profileName}...)
}

func PackReposPathOf(reposPath string) string {
	path := strings.NewReplacer("_", "__", "/", "_").Replace(reposPath)
	return filepath.Join(VimVoltOptDir(), path)
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

func VimExecutable() (string, error) {
	var vim string
	if vim = os.Getenv("VOLT_VIM"); vim != "" {
		return vim, nil
	}
	exeName := "vim"
	if runtime.GOOS == "windows" {
		exeName = "vim.exe"
	}
	return exec.LookPath(exeName)
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

func VimVoltOptDir() string {
	return filepath.Join(VimDir(), "pack", "volt", "opt")
}

func VimVoltStartDir() string {
	return filepath.Join(VimDir(), "pack", "volt", "start")
}

func BuildInfoJSON() string {
	return filepath.Join(VimVoltDir(), "build-info.json")
}

func BundledPlugConf() string {
	return filepath.Join(VimVoltStartDir(), "system", "plugin", "bundled_plugconf.vim")
}

func LookUpVimrcOrGvimrc() []string {
	return append(LookUpVimrc(), LookUpGvimrc()...)
}

func LookUpVimrc() []string {
	var vimrcPaths []string
	if runtime.GOOS == "windows" {
		vimrcPaths = []string{
			filepath.Join(HomeDir(), "_vimrc"),
			filepath.Join(VimDir(), "vimrc"),
		}
	} else {
		vimrcPaths = []string{
			filepath.Join(HomeDir(), ".vimrc"),
			filepath.Join(VimDir(), "vimrc"),
		}
	}
	for i := 0; i < len(vimrcPaths); {
		if !Exists(vimrcPaths[i]) {
			vimrcPaths = append(vimrcPaths[:i], vimrcPaths[i+1:]...)
			continue
		}
		i++
	}
	return vimrcPaths
}

func LookUpGvimrc() []string {
	var gvimrcPaths []string
	if runtime.GOOS == "windows" {
		gvimrcPaths = []string{
			filepath.Join(HomeDir(), "_gvimrc"),
			filepath.Join(VimDir(), "gvimrc"),
		}
	} else {
		gvimrcPaths = []string{
			filepath.Join(HomeDir(), ".gvimrc"),
			filepath.Join(VimDir(), "gvimrc"),
		}
	}
	for i := 0; i < len(gvimrcPaths); {
		if !Exists(gvimrcPaths[i]) {
			gvimrcPaths = append(gvimrcPaths[:i], gvimrcPaths[i+1:]...)
			continue
		}
		i++
	}
	return gvimrcPaths
}

func Exists(path string) bool {
	_, err := os.Lstat(path)
	return !os.IsNotExist(err)
}
