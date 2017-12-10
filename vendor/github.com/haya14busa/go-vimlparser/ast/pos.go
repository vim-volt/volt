package ast

import "fmt"

// Pos represents node position.
type Pos struct {
	Offset int // offset, starting at 0
	Line   int // line number, starting at 1
	Column int // column number, starting at 1 (byte count)

	// Should I support Filename?
	Filename string // filename, if any
}

// String returns a string in one of several forms:
//
//	file:line:column    valid position with file name
//	line:column         valid position without file name
//
func (pos Pos) String() string {
	s := pos.Filename
	if s != "" {
		s += ":"
	}
	s += fmt.Sprintf("%d:%d", pos.Line, pos.Column)
	return s
}
