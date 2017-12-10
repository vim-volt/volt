package vimlparser

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"runtime/debug"
	"strings"
	"testing"
)

func recovert(t testing.TB) {
	if r := recover(); r != nil {
		t.Errorf("Recovered: %v\n%s", r, debug.Stack())
	}
}

func TestNewStringReader(t *testing.T) {
	defer recovert(t)
	r := NewStringReader([]string{})
	if !r.eof() {
		t.Error("NewStringReader should call __init__ func to initialize")
	}
}

func TestStringReader___init__(t *testing.T) {
	defer recovert(t)
	tests := []struct {
		in  []string
		buf string
	}{
		{in: []string{}, buf: ""},
		{in: []string{""}, buf: "<EOL>"},
		{in: []string{"let x = 1"}, buf: "let x = 1<EOL>"},
		{in: []string{"let x = 1", "let y = x"}, buf: "let x = 1<EOL>let y = x<EOL>"},
		{in: []string{"let x =", `\ 1`}, buf: "let x = 1<EOL>"},
		{in: []string{"あいうえお"}, buf: "あいうえお<EOL>"},
	}
	for _, tt := range tests {
		r := &StringReader{}
		r.__init__(tt.in)
		if got := strings.Join(r.buf, ""); got != tt.buf {
			t.Errorf("StringReader.__init__(%v).buf == %v, want %v", tt.in, got, tt.buf)
		}
	}
}

func TestNewVimLParser(t *testing.T) {
	defer recovert(t)
	NewVimLParser(false).parse(NewStringReader([]string{}))
}

func TestVimLParser_parse_empty(t *testing.T) {
	defer recovert(t)
	ins := [][]string{
		{},
		{""},
		{"", ""},
	}
	for _, in := range ins {
		NewVimLParser(false).parse(NewStringReader(in))
	}
}

func TestVimLParser_parse(t *testing.T) {
	defer recovert(t)
	tests := []struct {
		in   []string
		want string
	}{
		{[]string{`" comment`}, "; comment"},
		{[]string{`let x = 1`}, "(let = x 1)"},
		{[]string{`call F(x, y, z)`}, "(call (F x y z))"},
	}
	for _, tt := range tests {
		c := NewCompiler()
		n := NewVimLParser(false).parse(NewStringReader(tt.in))
		if got := c.compile(n).([]string); strings.Join(got, "\n") != tt.want {
			t.Errorf("c.compile(p.parse(%v)) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

const basePkg = "github.com/haya14busa/go-vimlparser/go"

var skipTests = map[string]bool{
	"test_xxx_colonsharp": true,
}

func TestVimLParser_parse_compile(t *testing.T) {
	p, err := build.Default.Import(basePkg, "", build.FindOnly)
	if err != nil {
		t.Fatal(err)
	}
	testDir := path.Join(filepath.Dir(p.Dir), "test")
	// t.Error(p, err)
	vimfiles, err := filepath.Glob(testDir + "/test*.vim")
	if err != nil {
		t.Fatal(err)
	}
	for _, vimfile := range vimfiles {
		// t.Log(vimfile)

		ext := path.Ext(vimfile)
		base := vimfile[:len(vimfile)-len(ext)]
		okfile := base + ".ok"

		if b, ok := skipTests[path.Base(base)]; ok && b {
			t.Logf("Skip %v", path.Base(base))
			continue
		}

		in, err := readlines(vimfile)
		want, err := readlines(okfile)
		if err != nil {
			t.Error(err)
			continue
		}

		testFiles(t, path.Base(vimfile), in, want)
	}

}

func testFiles(t *testing.T, file string, in, want []string) {
	defer func() {
		if r := recover(); r != nil {
			err := strings.Trim(fmt.Sprintf("%s", r), "\n")
			w := strings.Trim(strings.Join(want, "\n"), "\n")
			if err != w {
				t.Log("===")
				t.Log("got :", err)
				t.Log("want:", w)
				// t.Log(w)
				t.Errorf("%v: Recovered: %v\n%s", file, r, debug.Stack())
				// t.Errorf("%v: Recovered: %v", file, r)
			}
		}
	}()

	neovim := false
	if strings.Contains(file, "test_neo") {
		neovim = true
	}

	r := NewStringReader(in)
	p := NewVimLParser(neovim)
	c := NewCompiler()
	got := c.compile(p.parse(r)).([]string)

	if g, w := strings.Trim(strings.Join(got, "\n"), "\n"), strings.Trim(strings.Join(want, "\n"), "\n"); g != w {
		t.Errorf("%v: got %v\nwant %v", file, g, w)
	}
}

func readlines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return strings.Split(string(b), "\n"), nil
}

func TestVimLParser_VimLParser(t *testing.T) {
	testParseVimLParser(t)
}

func BenchmarkVimLParser_VimLParser(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testParseVimLParser(b)
	}
}

func testParseVimLParser(t testing.TB) {
	defer recovert(t)
	p, err := build.Default.Import(basePkg, "", build.FindOnly)
	if err != nil {
		t.Fatal(err)
	}
	file := path.Join(filepath.Dir(p.Dir), "autoload", "vimlparser.vim")
	lines, err := readlines(file)
	if err != nil {
		t.Fatal(err)
	}
	c := NewCompiler()
	n := NewVimLParser(false).parse(NewStringReader(lines))
	c.compile(n)
}

func TestVimLParser_offset(t *testing.T) {
	defer recovert(t)
	const src = `

function! F()
  let x =
\ 1

  let x = "
  \1
	\2 <- tab
  \3 マルチバイト
  \4"
endfunction

" END
`
	const want = `function! F()
  let x =
\ 1

  let x = "
  \1
	\2 <- tab
  \3 マルチバイト
  \4"
endfunction`

	n := NewVimLParser(false).parse(NewStringReader(strings.Split(src, "\n")))
	f := n.body[0]
	start := f.pos
	end := f.endfunction.pos
	endfExArg := f.endfunction.ea
	end.offset += endfExArg.argpos.offset - endfExArg.cmdpos.offset

	if got := src[start.offset:end.offset]; got != want {
		t.Errorf("got:\n%v\nwant:\n%v", got, want)
	}
}
