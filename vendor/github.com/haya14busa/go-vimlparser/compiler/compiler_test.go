package compiler

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/haya14busa/go-vimlparser"
)

var skipTests = map[string]bool{
	"test_xxx_colonsharp": true,
}

func TestCompiler_Compile(t *testing.T) {
	vimfiles, err := filepath.Glob("../test/test*.vim")
	if err != nil {
		t.Fatal(err)
	}
	for _, vimfile := range vimfiles {
		ext := path.Ext(vimfile)
		base := vimfile[:len(vimfile)-len(ext)]
		okfile := base + ".ok"

		if b, ok := skipTests[path.Base(base)]; ok && b {
			t.Logf("Skip %v", path.Base(base))
			continue
		}

		testFile(t, vimfile, okfile)
	}
}

const okErrPrefix = "vimlparser: "

func testFile(t *testing.T, file, okfile string) {
	opt := &vimlparser.ParseOption{Neovim: strings.Contains(file, "test_neo")}
	in, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	node, err := vimlparser.ParseFile(in, "", opt)
	if err != nil {
		if !strings.HasPrefix(err.Error(), okErrPrefix) {
			t.Error(err)
		} else {
			return
		}
	}
	w := new(bytes.Buffer)
	if err := Compile(w, node); err != nil {
		t.Error(err)
	}

	b, err := ioutil.ReadAll(w)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Trim(string(b), "\n")

	b2, err := ioutil.ReadFile(okfile)
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Trim(string(b2), "\n")

	if got != want {
		t.Errorf("%v:\ngot:\n%v\nwant:\n%v", file, got, want)
	}
}

func BenchmarkCompiler_Compile(b *testing.B) {
	opt := &vimlparser.ParseOption{Neovim: false}
	in, err := os.Open("../autoload/vimlparser.vim")
	if err != nil {
		b.Fatal(err)
	}
	defer in.Close()
	node, err := vimlparser.ParseFile(in, "", opt)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		if err := Compile(ioutil.Discard, node); err != nil {
			b.Error(err)
		}
	}
}
