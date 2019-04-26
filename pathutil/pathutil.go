package pathutil

import (
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var rxReposPath = regexp.MustCompile(
	// scheme
	`^((?:https?|git)://)?` +
		// host
		`(?:([^/]+)/)?` +
		// user
		`(?:([^/]+)/)` +
		// name
		`([^/]+?)` +
		// trailing garbages
		`(?:\.git)?(/?)$`,
)

// NormalizeRepos normalizes name into the following forms into ReposPath:
// 1. user/name[.git]
// 2. github.com/user/name[.git]
// 3. [git|http|https]://github.com/user/name[.git][/]
func NormalizeRepos(rawReposPath string) (ReposPath, error) {
	p := filepath.ToSlash(rawReposPath)
	m := rxReposPath.FindStringSubmatch(p)
	if len(m) == 0 {
		return "", errors.New("invalid format of repository: " + rawReposPath)
	}
	if m[2] == "" {
		m[2] = "github.com"
	}
	disallowSlash := m[1] == ""
	if disallowSlash && m[5] == "/" {
		return "", errors.New("invalid format of repository: " + rawReposPath)
	}
	hostUserName := m[2:5]
	return ReposPath(strings.Join(hostUserName, "/")), nil
}

// ReposPath is string of "{site}/{user}/{repos}"
type ReposPath string

// ReposPathList is []ReposPath
type ReposPathList []ReposPath

func (path *ReposPath) String() string {
	return string(*path)
}

// Strings returns []string values.
func (list ReposPathList) Strings() []string {
	result := make([]string, 0, len(list))
	for i := range list {
		result = append(result, string(list[i]))
	}
	return result
}

// NormalizeLocalRepos normalizes name into ReposPath.
// If name does not contain "/", it is ReposPath("localhost/local/" + name).
// Otherwise same as NormalizeRepos(name).
func NormalizeLocalRepos(name string) (ReposPath, error) {
	if !strings.Contains(name, "/") {
		return ReposPath("localhost/local/" + name), nil
	}
	return NormalizeRepos(name)
}

// HomeDir detects HOME path.
// If HOME environment variable is not set,
// use USERPROFILE environment variable instead.
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

// VoltPath returns fullpath of "$HOME/volt".
func VoltPath() string {
	path := os.Getenv("VOLTPATH")
	if path != "" {
		return path
	}
	return filepath.Join(HomeDir(), "volt")
}

// FullPath returns fullpath of ReposPath.
func (path ReposPath) FullPath() string {
	reposList := strings.Split(filepath.ToSlash(path.String()), "/")
	paths := make([]string, 0, len(reposList)+2)
	paths = append(paths, VoltPath())
	paths = append(paths, "repos")
	paths = append(paths, reposList...)
	return filepath.Join(paths...)
}

// CloneURL returns string "https://{reposPath}".
func (path ReposPath) CloneURL() string {
	return "https://" + filepath.ToSlash(path.String())
}

// Plugconf returns fullpath of plugconf.
func (path ReposPath) Plugconf() string {
	filenameList := strings.Split(filepath.ToSlash(path.String()+".vim"), "/")
	paths := make([]string, 0, len(filenameList)+2)
	paths = append(paths, VoltPath())
	paths = append(paths, "plugconf")
	paths = append(paths, filenameList...)
	return filepath.Join(paths...)
}

// ProfileVimrc is the basename of profile vimrc.
const ProfileVimrc = "vimrc.vim"

// ProfileGvimrc is the basename of profile gvimrc.
const ProfileGvimrc = "gvimrc.vim"

// Vimrc is the basename of vimrc in ~/.vim
const Vimrc = "vimrc"

// Gvimrc is the basename of gvimrc in ~/.vim
const Gvimrc = "gvimrc"

// RCDir returns fullpath of "$HOME/volt/rc/{profileName}"
func RCDir(profileName string) string {
	return filepath.Join([]string{VoltPath(), "rc", profileName}...)
}

var packer = strings.NewReplacer("_", "__", "/", "_")
var unpacker1 = strings.NewReplacer("_", "/")
var unpacker2 = strings.NewReplacer("//", "_")

// EncodeToPlugDirName encodes path to directory name.
// The directory name is: ~/.vim/pack/volt/opt/{name}
func (path ReposPath) EncodeToPlugDirName() string {
	p := packer.Replace(path.String())
	return filepath.Join(VimVoltOptDir(), p)
}

// DecodeReposPath decodes name to repos path.
// name is directory name: ~/.vim/pack/volt/opt/{name}
func DecodeReposPath(name string) ReposPath {
	name = filepath.Base(name)
	return ReposPath(unpacker2.Replace(unpacker1.Replace(name)))
}

// LockJSON returns fullpath of "$HOME/volt/lock.json".
func LockJSON() string {
	return filepath.Join(VoltPath(), "lock.json")
}

// ConfigTOML returns fullpath of "$HOME/volt/config.toml".
func ConfigTOML() string {
	return filepath.Join(VoltPath(), "config.toml")
}

// TrxLock returns fullpath of "$HOME/volt/trx.lock".
func TrxLock() string {
	return filepath.Join(VoltPath(), "trx.lock")
}

// TempDir returns fullpath of "$HOME/tmp".
func TempDir() string {
	return filepath.Join(VoltPath(), "tmp")
}

// VimExecutable detects vim executable path.
// If VOLT_VIM environment variable is set, use it.
// Otherwise look up "vim" binary from PATH.
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

// VimDir returns the following fullpath:
//   Windows: $HOME/vimfiles
//   Other: $HOME/.vim
func VimDir() string {
	vimdir := ".vim"
	if runtime.GOOS == "windows" {
		vimdir = "vimfiles"
	}
	return filepath.Join(HomeDir(), vimdir)
}

// VimVoltDir returns "(vim dir)/pack/volt".
func VimVoltDir() string {
	return filepath.Join(VimDir(), "pack", "volt")
}

// VimVoltOptDir returns "(vim dir)/pack/volt/opt".
func VimVoltOptDir() string {
	return filepath.Join(VimDir(), "pack", "volt", "opt")
}

// VimVoltStartDir returns "(vim dir)/pack/volt/start".
func VimVoltStartDir() string {
	return filepath.Join(VimDir(), "pack", "volt", "start")
}

// BuildInfoJSON returns "(vim dir)/pack/volt/build-info.json".
func BuildInfoJSON() string {
	return filepath.Join(VimVoltDir(), "build-info.json")
}

// BundledPlugConf returns "(vim dir)/pack/volt/start/system/plugin/bundled_plugconf.vim".
func BundledPlugConf() string {
	return filepath.Join(VimVoltStartDir(), "system", "plugin", "bundled_plugconf.vim")
}

// LookUpVimrc looks up vimrc path from the following candidates:
//   Windows  : $HOME/_vimrc
//              (vim dir)/vimrc
//   Otherwise: $HOME/.vimrc
//              (vim dir)/vimrc
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

// LookUpGvimrc looks up gvimrc path from the following candidates:
//   Windows  : $HOME/_gvimrc
//              (vim dir)/gvimrc
//   Otherwise: $HOME/.gvimrc
//              (vim dir)/gvimrc
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

// Exists returns true if path exists, otherwise returns false.
// Existence is checked by os.Lstat().
func Exists(path string) bool {
	_, err := os.Lstat(path)
	return !os.IsNotExist(err)
}
