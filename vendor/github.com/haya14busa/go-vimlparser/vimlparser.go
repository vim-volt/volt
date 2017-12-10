package vimlparser

import (
	"bufio"
	"fmt"
	"io"

	"github.com/haya14busa/go-vimlparser/ast"
	internal "github.com/haya14busa/go-vimlparser/go"
)

// ErrVimlParser represents VimLParser error.
type ErrVimlParser struct {
	Filename string
	Offset   int
	Line     int
	Column   int
	Msg      string
}

func (e *ErrVimlParser) Error() string {
	if e.Filename != "" {
		return fmt.Sprintf("%v:%d:%d: vimlparser: %v", e.Filename, e.Line, e.Column, e.Msg)
	}
	return fmt.Sprintf("vimlparser: %v: line %d col %d", e.Msg, e.Line, e.Column)
}

// ParseOption is option for Parse().
type ParseOption struct {
	Neovim bool
}

// ParseFile parses Vim script.
// filename can be empty.
func ParseFile(r io.Reader, filename string, opt *ParseOption) (node *ast.File, err error) {
	defer func() {
		if r := recover(); r != nil {
			node = nil
			if e, ok := r.(*internal.ParseError); ok {
				err = &ErrVimlParser{
					Filename: filename,
					Offset:   e.Offset,
					Line:     e.Line,
					Column:   e.Column,
					Msg:      e.Msg,
				}
			} else {
				err = fmt.Errorf("%v", r)
			}
			// log.Printf("%s", debug.Stack())
		}
	}()
	lines := readlines(r)
	reader := internal.NewStringReader(lines)
	neovim := false
	if opt != nil {
		neovim = opt.Neovim
	}
	node = internal.NewVimLParser(neovim).Parse(reader, filename).(*ast.File)
	return
}

// ParseExpr parses Vim script expression.
func ParseExpr(r io.Reader) (node ast.Expr, err error) {
	defer func() {
		if r := recover(); r != nil {
			node = nil
			if e, ok := r.(*internal.ParseError); ok {
				err = &ErrVimlParser{
					Offset: e.Offset,
					Line:   e.Line,
					Column: e.Column,
					Msg:    e.Msg,
				}
			} else {
				err = fmt.Errorf("%v", r)
			}
			// log.Printf("%s", debug.Stack())
		}
	}()
	lines := readlines(r)
	reader := internal.NewStringReader(lines)
	p := internal.NewExprParser(reader)
	node = p.Parse()
	return
}

func readlines(r io.Reader) []string {
	lines := []string{}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}
