package subcmd

import (
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"regexp"
	"strconv"
)

// This variable is not constant for testing (to change it temporarily)
var voltVersion = "v0.3.6-alpha"

func init() {
	cmdMap["version"] = &versionCmd{}
}

type versionCmd struct {
	helped bool
}

func (cmd *versionCmd) ProhibitRootExecution(args []string) bool { return false }

func (cmd *versionCmd) FlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Print(`
Usage
  volt version [-help]

Description
  Show current version of volt.` + "\n\n")
		//fmt.Println("Options")
		//fs.PrintDefaults()
		fmt.Println()
		cmd.helped = true
	}
	return fs
}

func (cmd *versionCmd) Run(args []string) *Error {
	fs := cmd.FlagSet()
	fs.Parse(args)
	if cmd.helped {
		return nil
	}

	fmt.Printf("volt version: %s\n", voltVersion)
	return nil
}

// [major, minor, patch, alphaBeta]
type versionInfo []int

const (
	suffixAlpha  = 1
	suffixBeta   = 2
	suffixStable = 9
)

var rxVersion = regexp.MustCompile(`^v?([0-9]+)\.([0-9]+)(?:\.([0-9]+))?(-alpha|-beta)?`)

func voltVersionInfo() versionInfo {
	// parseVersion(voltVersionInfo) must not return non-nil error!
	voltVersionInfo, err := parseVersion(voltVersion)
	if err != nil {
		panic(err)
	}
	return voltVersionInfo
}

func compareVersion(v1, v2 versionInfo) int {
	for i := 0; i < 4; i++ {
		if v1[i] > v2[i] {
			return 1
		} else if v1[i] < v2[i] {
			return -1
		}
	}
	return 0
}

func parseVersion(ver string) (versionInfo, error) {
	m := rxVersion.FindStringSubmatch(ver)
	if len(m) == 0 {
		return nil, errors.New("version number format is invalid: " + ver)
	}
	info := make(versionInfo, 0, 4)
	for i := 1; i <= 3 && m[i] != ""; i++ {
		n, err := strconv.Atoi(m[i])
		if err != nil {
			return nil, err
		}
		info = append(info, n)
	}
	if len(info) == 2 {
		info = append(info, 0)
	}
	switch m[4] {
	case "":
		info = append(info, suffixStable)
	case "-alpha":
		info = append(info, suffixAlpha)
	case "-beta":
		info = append(info, suffixBeta)
	}
	return info, nil
}
