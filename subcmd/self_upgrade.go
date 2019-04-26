package subcmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"github.com/vim-volt/volt/httputil"
	"github.com/vim-volt/volt/logger"
)

func init() {
	cmdMap["self-upgrade"] = &selfUpgradeCmd{}
}

type selfUpgradeCmd struct {
	helped bool
	check  bool
}

func (cmd *selfUpgradeCmd) ProhibitRootExecution(args []string) bool { return true }

func (cmd *selfUpgradeCmd) FlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Print(`
Usage
  volt self-upgrade [-help] [-check]

Description
    Upgrade to the latest volt command, or if -check was given, it only checks the newer version is available.` + "\n\n")
		//fmt.Println("Options")
		//fs.PrintDefaults()
		fmt.Println()
		cmd.helped = true
	}
	fs.BoolVar(&cmd.check, "check", false, "only checks the newer version is available")
	return fs
}

func (cmd *selfUpgradeCmd) Run(args []string) *Error {
	err := cmd.parseArgs(args)
	if err == ErrShowedHelp {
		return nil
	}
	if err != nil {
		return &Error{Code: 10, Msg: "Failed to parse args: " + err.Error()}
	}

	if ppidStr := os.Getenv("VOLT_SELF_UPGRADE_PPID"); ppidStr != "" {
		if err = cmd.doCleanUp(ppidStr); err != nil {
			return &Error{Code: 11, Msg: "Failed to clean up old binary: " + err.Error()}
		}
	} else {
		latestURL := "https://api.github.com/repos/vim-volt/volt/releases/latest"
		if err = cmd.doSelfUpgrade(latestURL); err != nil {
			return &Error{Code: 12, Msg: "Failed to self-upgrade: " + err.Error()}
		}
	}

	return nil
}

func (cmd *selfUpgradeCmd) parseArgs(args []string) error {
	fs := cmd.FlagSet()
	fs.Parse(args)
	if cmd.helped {
		return ErrShowedHelp
	}
	return nil
}

func (cmd *selfUpgradeCmd) doCleanUp(ppidStr string) error {
	ppid, err := strconv.Atoi(ppidStr)
	if err != nil {
		return errors.Wrap(err, "failed to parse VOLT_SELF_UPGRADE_PPID")
	}

	// Wait until the parent process exits
	if died := cmd.waitUntilParentExits(ppid); !died {
		return errors.Errorf("parent pid (%s) is keeping alive for long time", ppidStr)
	}

	// Remove old binary
	voltExe, err := cmd.getExecutablePath()
	if err != nil {
		return err
	}
	return os.Remove(voltExe + ".old")
}

func (cmd *selfUpgradeCmd) waitUntilParentExits(pid int) bool {
	fib := []int{1, 1, 2, 3, 5, 8, 13} // 33 second
	for i := 0; i < len(fib); i++ {
		if !cmd.processIsAlive(pid) {
			return true
		}
		time.Sleep(time.Duration(fib[i]) * time.Second)
	}
	return false
}

func (*selfUpgradeCmd) processIsAlive(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

type latestRelease struct {
	TagName string `json:"tag_name"`
	Body    string `json:"body"`
	Assets  []releaseAsset
}

type releaseAsset struct {
	BrowserDownloadURL string `json:"browser_download_url"`
	Name               string `json:"name"`
}

func (cmd *selfUpgradeCmd) doSelfUpgrade(latestURL string) error {
	// Check the latest binary info
	release, err := cmd.checkLatest(latestURL)
	if err != nil {
		return err
	}
	logger.Debugf("tag_name = %q", release.TagName)
	tagNameVer, err := parseVersion(release.TagName)
	if err != nil {
		return err
	}
	if compareVersion(tagNameVer, voltVersionInfo()) <= 0 {
		logger.Info("No updates were found.")
		return nil
	}
	logger.Infof("Found update: %s -> %s", voltVersion, release.TagName)

	// Show release note
	fmt.Println("---")
	fmt.Println(release.Body)
	fmt.Println("---")

	if cmd.check {
		return nil
	}

	// Download the latest binary as "volt[.exe].latest"
	voltExe, err := cmd.getExecutablePath()
	if err != nil {
		return err
	}
	latestFile, err := os.OpenFile(voltExe+".latest", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	err = cmd.download(latestFile, release)
	latestFile.Close()
	if err != nil {
		return err
	}

	// Rename dir/volt[.exe] to dir/volt[.exe].old
	// NOTE: Windows can rename running executable file
	if err := os.Rename(voltExe, voltExe+".old"); err != nil {
		return err
	}

	// Rename dir/volt[.exe].latest to dir/volt[.exe]
	if err := os.Rename(voltExe+".latest", voltExe); err != nil {
		return err
	}

	// Spawn dir/volt[.exe] with env "VOLT_SELF_UPGRADE_PPID={pid}"
	voltCmd := exec.Command(voltExe, "self-upgrade")
	if err = voltCmd.Start(); err != nil {
		return err
	}
	return nil
}

func (*selfUpgradeCmd) getExecutablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}

func (*selfUpgradeCmd) checkLatest(url string) (*latestRelease, error) {
	content, err := httputil.GetContent(url)
	if err != nil {
		return nil, err
	}
	var release latestRelease
	if err = json.Unmarshal(content, &release); err != nil {
		return nil, err
	}
	return &release, nil
}

func (*selfUpgradeCmd) download(w io.Writer, release *latestRelease) error {
	suffix := runtime.GOOS + "-" + runtime.GOARCH
	for i := range release.Assets {
		// e.g.: Name = "volt-v0.1.2-linux-amd64"
		if strings.HasSuffix(release.Assets[i].Name, suffix) {
			r, err := httputil.GetContentReader(release.Assets[i].BrowserDownloadURL)
			if err != nil {
				return err
			}
			defer r.Close()
			if _, err = io.Copy(w, r); err != nil {
				return err
			}
			break
		}
	}
	return nil
}
