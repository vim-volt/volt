package usecase

import (
	"errors"
	"regexp"
	"strconv"
)

// NOTE: this is not constant for testing (to change it temporarily)
var voltVersion = "v0.3.6-alpha"

// VersionString is current version string
func VersionString() string {
	return voltVersion
}

// VersionInfo is [major, minor, patch, alphaBeta]
type VersionInfo []int

const (
	suffixAlpha  = 1
	suffixBeta   = 2
	suffixStable = 9
)

var rxVersion = regexp.MustCompile(`^v?([0-9]+)\.([0-9]+)(?:\.([0-9]+))?(-alpha|-beta)?`)

// Version returns versions [major, minor, patch, alphaBeta]
func Version() VersionInfo {
	// parseVersion(voltVersion) must not return non-nil error!
	voltVersionInfo, err := ParseVersion(voltVersion)
	if err != nil {
		panic(err)
	}
	return voltVersionInfo
}

// CompareVersion compares two versions.
// and returns negative if v1 < v2, or positive if v1 > v2, or 0 if v1 == v2.
func CompareVersion(v1, v2 VersionInfo) int {
	for i := 0; i < 4; i++ {
		if v1[i] > v2[i] {
			return 1
		} else if v1[i] < v2[i] {
			return -1
		}
	}
	return 0
}

// ParseVersion parses version string
func ParseVersion(ver string) (VersionInfo, error) {
	m := rxVersion.FindStringSubmatch(ver)
	if len(m) == 0 {
		return nil, errors.New("version number format is invalid: " + ver)
	}
	info := make(VersionInfo, 0, 4)
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
