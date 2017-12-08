package ast_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/haya14busa/go-vimlparser"
	"github.com/haya14busa/go-vimlparser/ast"
)

func TestInspect(t *testing.T) {
	match, err := filepath.Glob("../test/test_*.vim")
	if err != nil {
		t.Fatal(err)
	}
	okErr := "vimlparser:"
	match = append(match, "../autoload/vimlparser.vim")
	match = append(match, "../go/gocompiler.vim")
	for _, filename := range match {
		if err := checkInspect(t, filename); err != nil && !strings.HasPrefix(err.Error(), okErr) {
			t.Errorf("%s: %v", filename, err)
		}
	}
}

func checkInspect(t testing.TB, filename string) error {
	in, err := os.Open(filename)
	if err != nil {
		t.Error(err)
	}
	defer in.Close()
	f, err := vimlparser.ParseFile(in, "", nil)
	if err != nil {
		return err
	}
	ast.Inspect(f, func(n ast.Node) bool {
		// do nothing
		return true
	})

	return nil
}
