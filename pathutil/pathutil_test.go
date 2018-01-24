package pathutil

import "testing"

func TestNormalizeRepos(t *testing.T) {
	var tests = []struct {
		in  string
		out ReposPath
	}{
		{"user/name", ReposPath("github.com/user/name")},
		{"user/name.git", ReposPath("github.com/user/name")},
		{"github.com/user/name", ReposPath("github.com/user/name")},
		{"github.com/user/name.git", ReposPath("github.com/user/name")},
		{"https://github.com/user/name", ReposPath("github.com/user/name")},
		{"https://github.com/user/name.git", ReposPath("github.com/user/name")},
		{"https://github.com/user/name/", ReposPath("github.com/user/name")},
		{"https://github.com/user/name.git/", ReposPath("github.com/user/name")},
		{"http://github.com/user/name", ReposPath("github.com/user/name")},
		{"http://github.com/user/name.git", ReposPath("github.com/user/name")},
		{"http://github.com/user/name/", ReposPath("github.com/user/name")},
		{"http://github.com/user/name.git/", ReposPath("github.com/user/name")},
		{"git://github.com/user/name", ReposPath("github.com/user/name")},
		{"git://github.com/user/name.git", ReposPath("github.com/user/name")},
		{"git://github.com/user/name/", ReposPath("github.com/user/name")},
		{"git://github.com/user/name.git/", ReposPath("github.com/user/name")},
		{"localhost/local/name", ReposPath("localhost/local/name")},
		{"localhost/local/name.git", ReposPath("localhost/local/name")},
	}
	for _, tt := range tests {
		result, err := NormalizeRepos(tt.in)
		if err != nil {
			t.Errorf("in:%s, err:%s", tt.in, err.Error())
		}
		if result != tt.out {
			t.Errorf("in:%s, got:%s, expected:%s", tt.in, result, tt.out)
		}
	}
}

func TestNormalizeReposError(t *testing.T) {
	// protocols other than git, http, https
	var tests = []string{
		"ftp://github.com/user/name",
		"ftp://github.com/user/name.git",
		"user/name/",
		"github.com/user/name/",
	}
	for _, tt := range tests {
		_, err := NormalizeRepos(tt)
		if err == nil {
			t.Errorf("in:%s -> expected error but no error", tt)
		}
	}
}
