package vimlparser

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/haya14busa/go-vimlparser/compiler"
)

func TestParseFile_can_parse(t *testing.T) {
	match, err := filepath.Glob("test/test_*.vim")
	if err != nil {
		t.Fatal(err)
	}
	okErr := "vimlparser:"
	match = append(match, "autoload/vimlparser.vim")
	match = append(match, "go/gocompiler.vim")
	for _, filename := range match {
		if err := checkParse(t, filename); err != nil && !strings.HasPrefix(err.Error(), okErr) {
			t.Errorf("%s: %v", filename, err)
		}
	}
}

func checkParse(t testing.TB, filename string) error {
	in, err := os.Open(filename)
	if err != nil {
		t.Error(err)
	}
	defer in.Close()
	_, err = ParseFile(in, "", nil)
	return err
}

func BenchmarkParseFile(b *testing.B) {
	filename := "autoload/vimlparser.vim"
	for i := 0; i < b.N; i++ {
		checkParse(b, filename)
	}
}

func TestParseFile_error(t *testing.T) {
	want := "path/to/filename.go:1:1: vimlparser: E492: Not an editor command: hoge"
	_, err := ParseFile(strings.NewReader("hoge"), "path/to/filename.go", nil)
	if err != nil {

		if er, ok := err.(*ErrVimlParser); !ok {
			t.Errorf("Error type is %T, want %T", er, &ErrVimlParser{})
		}

		if got := err.Error(); want != got {
			t.Errorf("ParseFile(\"hoge\") = %v, want %v", got, want)
		}
	}
}

func TestParseExpr_Compile(t *testing.T) {
	node, err := ParseExpr(strings.NewReader("x + 1"))
	if err != nil {
		t.Fatal(err)
	}
	c := compiler.Compiler{Config: compiler.Config{Indent: "  "}}
	b := new(bytes.Buffer)
	if err := c.Compile(b, node); err != nil {
		t.Fatal(err)
	}
	if got, want := b.String(), "(+ x 1)"; got != want {
		t.Errorf("Compile(ParseExpr(\"x + 1\")) = %v, want %v", got, want)
	}
}

func TestParseExpr_Parser_err(t *testing.T) {
	want := "vimlparser: unexpected token: /: line 1 col 4"
	_, err := ParseExpr(strings.NewReader("1 // 2"))
	if err != nil {

		if er, ok := err.(*ErrVimlParser); !ok {
			t.Errorf("Error type is %T, want %T", er, &ErrVimlParser{})
		}

		if got := err.Error(); want != got {
			t.Errorf("ParseExpr(\"1 // 2\") = %v, want %v", got, want)
		}
	}
}
