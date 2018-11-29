package usecase

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"github.com/vim-volt/volt/httputil"
	"github.com/vim-volt/volt/logger"
)

// SelfUpgrade upgrades running volt binary if checkOnly = false.
// if checkOnly = true, only check the latest version and shows the information.
func SelfUpgrade(latestURL string, checkOnly bool) error {
	// Check the latest binary info
	release, err := checkLatest(latestURL)
	if err != nil {
		return err
	}
	logger.Debugf("tag_name = %q", release.TagName)
	tagNameVer, err := ParseVersion(release.TagName)
	if err != nil {
		return err
	}
	if CompareVersion(tagNameVer, Version()) <= 0 {
		logger.Info("No updates were found.")
		return nil
	}
	logger.Infof("Found update: %s -> %s", VersionString(), release.TagName)

	// Show release note
	fmt.Println("---")
	fmt.Println(release.Body)
	fmt.Println("---")

	if checkOnly {
		return nil
	}

	// Download the latest binary as "volt[.exe].latest"
	voltExe, err := getExecutablePath()
	if err != nil {
		return err
	}
	latestFile, err := os.OpenFile(voltExe+".latest", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	err = download(latestFile, release)
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

func download(w io.Writer, release *Release) error {
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

// checkLatest returns the latest release information.
func checkLatest(url string) (*Release, error) {
	content, err := httputil.GetContent(url)
	if err != nil {
		return nil, err
	}
	var release Release
	if err = json.Unmarshal(content, &release); err != nil {
		return nil, err
	}
	return &release, nil
}

func getExecutablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}

// Release has information about a volt release.
type Release struct {
	TagName string `json:"tag_name"`
	Body    string `json:"body"`
	Assets  []struct {
		BrowserDownloadURL string `json:"browser_download_url"`
		Name               string `json:"name"`
	}
}

// RemoveOldBinary removes old
func RemoveOldBinary(ppid int) error {
	// Wait until the parent process exits
	if died := waitUntilParentExits(ppid); !died {
		return errors.Errorf("parent pid (%d) is keeping alive for long time", ppid)
	}

	// Remove old binary
	voltExe, err := getExecutablePath()
	if err != nil {
		return err
	}
	return os.Remove(voltExe + ".old")
}

func waitUntilParentExits(pid int) bool {
	fib := []int{1, 1, 2, 3, 5, 8, 13} // 33 second
	for i := 0; i < len(fib); i++ {
		if !processIsAlive(pid) {
			return true
		}
		time.Sleep(time.Duration(fib[i]) * time.Second)
	}
	return false
}

func processIsAlive(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil
}
