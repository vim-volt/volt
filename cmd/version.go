package cmd

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

var voltVersion string = "v0.1.3-beta"

func Version(args []string) int {
	fmt.Printf("volt version: %s\n", voltVersion)

	return 0
}

// [major, minor, patch, alphaBeta]
type versionInfo []int

const (
	suffixAlpha  = 1
	suffixBeta   = 2
	suffixStable = 9
)

var rxVersion = regexp.MustCompile(`^v?([0-9]+)\.([0-9]+)(?:\.([0-9]+))?(-alpha|-beta)?`)
var voltVersionInfo versionInfo

func init() {
	var err error
	voltVersionInfo, err = parseVersion(voltVersion)
	if err != nil {
		panic(err)
	}
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
