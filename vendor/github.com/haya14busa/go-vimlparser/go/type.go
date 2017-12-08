package vimlparser

import (
	"fmt"
	"strings"
)

type ExArg struct {
	forceit      bool
	addr_count   int
	line1        int
	line2        int
	flags        int
	do_ecmd_cmd  string
	do_ecmd_lnum int
	append       int
	usefilter    bool
	amount       int
	regname      int
	force_bin    int
	read_edit    int
	force_ff     string // int
	force_enc    string // int
	bad_char     string // int
	linepos      *pos
	cmdpos       *pos
	argpos       *pos
	cmd          *Cmd
	modifiers    []interface{}
	range_       []interface{} // range -> range_
	argopt       map[string]interface{}
	argcmd       map[string]interface{}
}

type Cmd struct {
	name   string
	minlen int
	flags  string
	parser string
}

type VimNode struct {
	type_ int // type -> type_
	pos   *pos
	left  *VimNode
	right *VimNode
	cond  *VimNode
	rest  *VimNode
	list  []*VimNode
	rlist []*VimNode
	body  []*VimNode
	op    string
	str   string
	depth int
	value interface{}

	ea   *ExArg
	attr *FuncAttr

	endfunction *VimNode
	elseif      []*VimNode
	else_       *VimNode
	endif       *VimNode
	endwhile    *VimNode
	endfor      *VimNode
	endtry      *VimNode

	catch   []*VimNode
	finally *VimNode

	pattern string
	curly   bool
}

type FuncAttr struct {
	range_  bool
	abort   bool
	dict    bool
	closure bool
}

type lhs struct {
	left *VimNode
	list []*VimNode
	rest *VimNode
}

type pos struct {
	i      int
	lnum   int
	col    int
	offset int
}

// Node returns new VimNode.
func Node(type_ int) *VimNode {
	if type_ == -1 {
		return nil
	}
	return &VimNode{
		type_: type_,
		attr:  &FuncAttr{},
	}
}

type VimLParser struct {
	find_command_cache map[string]*Cmd
	reader             *StringReader
	context            []*VimNode
	ea                 *ExArg
	neovim             bool
}

func NewVimLParser(neovim bool) *VimLParser {
	obj := &VimLParser{}
	obj.find_command_cache = make(map[string]*Cmd)
	obj.__init__(neovim)
	return obj
}

func (self *VimLParser) __init__(neovim bool) {
	self.neovim = neovim
}

func (self *VimLParser) push_context(n *VimNode) {
	self.context = append([]*VimNode{n}, self.context...)
}

func (self *VimLParser) pop_context() {
	self.context = self.context[1:]
}

type ExprToken struct {
	type_ int
	value string
	pos   *pos
}

type ExprTokenizer struct {
	reader *StringReader
	cache  map[int][]interface{} // (int, *ExprToken)
}

func NewExprTokenizer(reader *StringReader) *ExprTokenizer {
	obj := &ExprTokenizer{}
	obj.cache = make(map[int][]interface{})
	obj.__init__(reader)
	return obj
}

func (self *ExprTokenizer) token(type_ int, value string, pos *pos) *ExprToken {
	return &ExprToken{type_: type_, value: value, pos: pos}
}

type ExprParser struct {
	reader    *StringReader
	tokenizer *ExprTokenizer
}

func NewExprParser(reader *StringReader) *ExprParser {
	obj := &ExprParser{}
	obj.__init__(reader)
	return obj
}

type LvalueParser struct {
	*ExprParser
}

func NewLvalueParser(reader *StringReader) *LvalueParser {
	obj := &LvalueParser{&ExprParser{}}
	obj.__init__(reader)
	return obj
}

type StringReader struct {
	i   int
	pos []pos
	buf []string
}

func NewStringReader(lines []string) *StringReader {
	obj := &StringReader{}
	obj.__init__(lines)
	return obj
}

func (self *StringReader) __init__(lines []string) {
	size := 0
	for _, l := range lines {
		size += len(l) + 1 // +1 for EOL
	}
	self.buf = make([]string, 0, size)
	self.pos = make([]pos, 0, size+1) // +1 for EOF
	var lnum = 0
	var offset = 0
	for lnum < len(lines) {
		var col = 0
		for _, r := range lines[lnum] {
			c := string(r)
			self.buf = append(self.buf, c)
			self.pos = append(self.pos, pos{lnum: lnum + 1, col: col + 1, offset: offset})
			col += len(c)
			offset += len(c)
		}
		for lnum+1 < len(lines) && viml_eqregh(lines[lnum+1], "^\\s*\\\\") {
			var skip = true
			col = 0
			for _, r := range lines[lnum+1] {
				c := string(r)
				if skip {
					if c == "\\" {
						skip = false
					}
				} else {
					self.buf = append(self.buf, c)
					self.pos = append(self.pos, pos{lnum: lnum + 2, col: col + 1})
				}
				col += len(c)
				offset += len(c)
			}
			lnum += 1
			offset += 1
		}
		self.buf = append(self.buf, "<EOL>")
		self.pos = append(self.pos, pos{lnum: lnum + 1, col: col + 1, offset: offset})
		lnum += 1
		offset += 1
	}
	// for <EOF>
	self.pos = append(self.pos, pos{lnum: lnum + 1, col: 0, offset: offset})
	self.i = 0
}

func (self *StringReader) getpos() *pos {
	p := self.pos[self.i]
	p.i = self.i
	return &p
}

type Compiler struct {
	indent []string
	lines  []string
}

func NewCompiler() *Compiler {
	obj := &Compiler{}
	obj.__init__()
	return obj
}

func (self *Compiler) __init__() {
	self.indent = []string{""}
	self.lines = []string{}
}

func (self *Compiler) out(f string, args ...interface{}) {
	if len(args) == 0 {
		if string(f[0]) == ")" {
			self.lines[len(self.lines)-1] += f
		} else {
			self.lines = append(self.lines, self.indent[0]+f)
		}
	} else {
		self.lines = append(self.lines, self.indent[0]+viml_printf(f, args...))
	}
}

func (self *Compiler) incindent(s string) {
	self.indent = append([]string{self.indent[0] + s}, self.indent...)
}

func (self *Compiler) decindent() {
	self.indent = self.indent[1:]
}

func (self *Compiler) compile_curlynameexpr(n *VimNode) string {
	return "{" + self.compile(n.value.(*VimNode)).(string) + "}"
}

func (self *Compiler) compile_list(n *VimNode) string {
	var value = func() []string {
		var ss []string
		for _, vval := range n.value.([]interface{}) {
			ss = append(ss, self.compile(vval.(*VimNode)).(string))
		}
		return ss
	}()
	if viml_empty(value) {
		return "(list)"
	} else {
		return viml_printf("(list %s)", viml_join(value, " "))
	}
}

func (self *Compiler) compile_curlyname(n *VimNode) string {
	return viml_join(func() []string {
		var ss []string
		for _, vval := range n.value.([]*VimNode) {
			ss = append(ss, self.compile(vval).(string))
		}
		return ss
	}(), "")
}

func (self *Compiler) compile_dict(n *VimNode) string {
	var value = []string{}
	for _, nn := range n.value.([]interface{}) {
		kv := nn.([]interface{})
		value = append(value, "("+self.compile(kv[0].(*VimNode)).(string)+" "+self.compile(kv[1].(*VimNode)).(string)+")")
	}
	if viml_empty(value) {
		return "(dict)"
	} else {
		return viml_printf("(dict %s)", strings.Join(value, " "))
	}
}

func (self *Compiler) compile_parenexpr(n *VimNode) string {
	return self.compile(n.value.(*VimNode)).(string)
}

type ParseError struct {
	Offset int
	Line   int
	Column int
	Msg    string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("vimlparser: %s: line %d col %d", e.Msg, e.Line, e.Column)
}

func Err(msg string, pos *pos) *ParseError {
	return &ParseError{Offset: pos.i, Line: pos.lnum, Column: pos.col, Msg: msg}
}
