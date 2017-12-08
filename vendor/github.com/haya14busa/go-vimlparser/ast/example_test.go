package ast_test

import (
	"fmt"
	"log"

	"strings"

	"github.com/haya14busa/go-vimlparser"
	"github.com/haya14busa/go-vimlparser/ast"
)

// This example demonstrates how to inspect the AST of a Go program.
func ExampleInspect() {
	// src is the input for which we want to inspect the AST.
	src := `
let s:c = 1.0
let X = F(3.14)*2 + s:c
`

	opt := &vimlparser.ParseOption{}
	f, err := vimlparser.ParseFile(strings.NewReader(src), "src.vim", opt)
	if err != nil {
		log.Fatal(err)
	}

	// Inspect the AST and print all identifiers and literals.
	ast.Inspect(f, func(n ast.Node) bool {
		var s string
		switch x := n.(type) {
		case *ast.BasicLit:
			s = x.Value
		case *ast.Ident:
			s = x.Name
		}
		if s != "" {
			fmt.Printf("%s:\t%s\n", n.Pos(), s)
		}
		return true
	})

	// output:
	// src.vim:2:5:	s:c
	// src.vim:2:11:	1.0
	// src.vim:3:5:	X
	// src.vim:3:9:	F
	// src.vim:3:11:	3.14
	// src.vim:3:17:	2
	// src.vim:3:21:	s:c
}
