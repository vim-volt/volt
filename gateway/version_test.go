package gateway

import "testing"

func TestCompareVersion(t *testing.T) {
	for _, tt := range []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0", "1.0.0", 0},
		{"0.1", "0.1.2", -1},
		{"0.1.10", "0.1.2", 1},
		{"0.2.0", "0.1.2", 1},
		{"1.2.0", "1.1.10", 1},
		{"0.1.3-alpha", "0.1.3-beta", -1},
		{"0.1.3-beta", "0.1.3", -1},
		{"0.1.3", "0.1.3-alpha", 1},
		{"1.0-alpha", "1.0-beta", -1},
		{"1.0-beta", "1.0", -1},
		{"1.0", "1.0-alpha", 1},
	} {
		for _, prefix := range []string{"", "v"} {
			v1 := parse(t, prefix+tt.v1)
			v2 := parse(t, prefix+tt.v2)

			if got := compareVersion(v1, v2); got != tt.expected {
				t.Errorf("compareVersion(%q = %v, %q = %v): expected %v but got %v", prefix+tt.v1, v1, prefix+tt.v2, v2, tt.expected, got)
			}

			if got := compareVersion(v2, v1); got != -tt.expected {
				t.Errorf("compareVersion(%q = %v, %q = %v): expected %v but got %v", prefix+tt.v2, v2, prefix+tt.v1, v1, -tt.expected, got)
			}
		}
	}
}

func parse(t *testing.T, ver string) versionInfo {
	vinfo, err := parseVersion(ver)
	if err != nil {
		t.Errorf("\"%s\" should be a version number but isn't: %s", ver, err.Error())
	}
	if len(vinfo) != 4 {
		t.Errorf("parseVersion(%q) returned invalid versionInfo: %q", ver, vinfo)
	}
	return vinfo
}
