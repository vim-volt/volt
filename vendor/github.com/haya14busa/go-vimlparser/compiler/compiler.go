// Package compiler provides compiler from Vim script AST into S-expression
// like format which is the same format as Compiler of vim-vimlparser.
// ref: "go/printer"
package compiler

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/haya14busa/go-vimlparser/ast"
	"github.com/haya14busa/go-vimlparser/token"
)

// Config for Compiler.
type Config struct {
	Indent string
}

// Compiler compiles Vim AST.
type Compiler struct {
	Config

	// Current state
	buffer *bytes.Buffer // raw compiler result
	indent int           // current indentation
}

// Compile compiles node and writes to writer.
func Compile(w io.Writer, node ast.Node) error {
	return (&Compiler{Config: Config{Indent: "  "}}).Compile(w, node)
}

// Compile compiles node and writes to writer.
func (c *Compiler) Compile(w io.Writer, node ast.Node) error {
	c.buffer = bytes.NewBuffer(make([]byte, 0))
	if err := c.compile(node); err != nil {
		return err
	}
	if _, err := io.Copy(w, c.buffer); err != nil {
		return err
	}
	return nil
}

func (c *Compiler) fprintln(f string, args ...interface{}) {
	c.fprint(f+"\n", args...)
}

func (c *Compiler) fprint(f string, args ...interface{}) {
	fmt.Fprint(c.buffer, strings.Repeat(c.Config.Indent, c.indent))
	fmt.Fprintf(c.buffer, f, args...)
}

func (c *Compiler) writeString(s string) {
	fmt.Fprint(c.buffer, s)
}

func (c *Compiler) trimLineBreak() {
	new := bytes.TrimRight(c.buffer.Bytes(), "\n")
	c.buffer.Reset()
	c.buffer.Write(new)
}

func (c *Compiler) compile(node interface{}) error {
	switch n := node.(type) {
	case *ast.File:
		return c.compileFile(n)
	case *ast.Comment:
		return c.compileComment(n)
	case ast.ExCommand:
		return c.compileExcommand(n)
	case []ast.Statement:
		for _, s := range n {
			if err := c.compile(s); err != nil {
				return err
			}
		}
	case ast.Expr:
		c.fprint(c.compileExpr(n))
	}
	return nil
}

func (c *Compiler) compileFile(node *ast.File) error {
	for _, stmt := range node.Body {
		if err := c.compile(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiler) compileComment(node *ast.Comment) error {
	c.fprintln(";%s", node.Text)
	return nil
}

func (c *Compiler) compileExcommand(node ast.ExCommand) error {
	switch n := node.(type) {
	case *ast.Excmd:
		c.compileExcmd(n)
	case *ast.Function:
		c.compileFunction(n)
	case *ast.DelFunction:
		c.compileDelfunction(n)
	case *ast.Return:
		c.compileReturn(n)
	case *ast.ExCall:
		c.compileExcall(n)
	case *ast.Let:
		c.compileLet(n)
	case *ast.UnLet:
		c.compileUnlet(n)
	case *ast.LockVar:
		c.compileLockvar(n)
	case *ast.UnLockVar:
		c.compileUnlockvar(n)
	case *ast.If:
		c.compileIf(n)
	case *ast.While:
		c.compileWhile(n)
	case *ast.For:
		c.compileFor(n)
	case *ast.Continue, *ast.Break:
		c.compileSingleCmd(n)
	case *ast.Try:
		c.compileTry(n)
	case *ast.Throw:
		c.compileThrow(n)
	case *ast.EchoCmd:
		c.compileEchocmd(n)
	case *ast.Echohl:
		c.compileEchohl(n)
	case *ast.Execute:
		c.compileExecute(n)
	case *ast.EndFor, *ast.EndIf, *ast.Finally, *ast.EndFunction, *ast.EndTry,
		*ast.EndWhile, *ast.Catch, *ast.Else, *ast.ElseIf:
		return fmt.Errorf("compileExcommand: unexpected Node: %v", n)
	}
	return nil
}

func (c *Compiler) compileExcmd(node *ast.Excmd) {
	c.fprintln(`(excmd "%s")`, escape(node.Command, `\"`))
}

func (c *Compiler) compileFunction(node *ast.Function) {
	c.fprint("(function (%s", c.compileExpr(node.Name))
	if len(node.Params) > 0 {
		c.fprint(" ")
		ps := make([]string, 0, len(node.Params))
		for _, p := range node.Params {
			if p.Name == token.DOTDOTDOT.String() {
				ps = append(ps, ". ...")
			} else {
				ps = append(ps, p.Name)
			}
		}
		c.fprint("%s", strings.Join(ps, " "))
	}
	c.fprintln(")")
	c.indent++
	c.compile(node.Body)
	c.trimLineBreak()
	c.writeString(")\n")
	c.indent--
}

func (c *Compiler) compileDelfunction(node *ast.DelFunction) {
	c.fprintln("(%s %s)", node.Cmd().Name, c.compileExpr(node.Name))
}

func (c *Compiler) compileReturn(node *ast.Return) {
	if node.Result != nil {
		c.fprintln("(%s %s)", node.Cmd().Name, c.compileExpr(node.Result))
	} else {
		c.fprintln("(%s)", node.Cmd().Name)
	}
}

func (c *Compiler) compileExcall(node *ast.ExCall) {
	c.fprintln("(%s %s)", node.Cmd().Name, c.compileExpr(node.FuncCall))
}

func (c *Compiler) compileLet(node *ast.Let) {
	cmd := node.Cmd().Name
	lhs := ""
	if node.Left != nil {
		lhs = c.compileExpr(node.Left)
	} else {
		ls := make([]string, 0, len(node.List))
		for _, n := range node.List {
			ls = append(ls, c.compileExpr(n))
		}
		rest := ""
		if node.Rest != nil {
			rest = " . " + c.compileExpr(node.Rest)
		}
		lhs = fmt.Sprintf("(%s%s)", strings.Join(ls, " "), rest)
	}
	rhs := c.compileExpr(node.Right)
	c.fprintln("(%s %s %s %s)", cmd, node.Op, lhs, rhs)
}

func (c *Compiler) compileUnlet(node *ast.UnLet) {
	cmd := node.Cmd().Name
	list := make([]string, 0, len(node.List))
	for _, n := range node.List {
		list = append(list, c.compileExpr(n))
	}
	c.fprintln("(%s %s)", cmd, strings.Join(list, " "))
}

func (c *Compiler) compileLockvar(node *ast.LockVar) {
	cmd := node.Cmd().Name
	list := make([]string, 0, len(node.List))
	for _, n := range node.List {
		list = append(list, c.compileExpr(n))
	}
	if node.Depth > 0 {
		c.fprintln("(%s %d %s)", cmd, node.Depth, strings.Join(list, " "))
	} else {
		c.fprintln("(%s %s)", cmd, strings.Join(list, " "))
	}
}

func (c *Compiler) compileUnlockvar(node *ast.UnLockVar) {
	cmd := node.Cmd().Name
	list := make([]string, 0, len(node.List))
	for _, n := range node.List {
		list = append(list, c.compileExpr(n))
	}
	if node.Depth > 0 {
		c.fprintln("(%s %d %s)", cmd, node.Depth, strings.Join(list, " "))
	} else {
		c.fprintln("(%s %s)", cmd, strings.Join(list, " "))
	}
}

func (c *Compiler) compileIf(node *ast.If) {
	cmd := node.Cmd().Name
	c.fprintln("(%s %s", cmd, c.compileExpr(node.Condition))
	c.indent++
	c.compile(node.Body)
	c.indent--
	for _, n := range node.ElseIf {
		c.fprintln(" %s %s", n.Cmd().Name, c.compileExpr(n.Condition))
		c.indent++
		c.compile(n.Body)
		c.indent--
	}
	if node.Else != nil {
		c.fprintln(" %s", node.Else.Cmd().Name)
		c.indent++
		c.compile(node.Else.Body)
		c.indent--
	}
	c.trimLineBreak()
	c.writeString(")\n")
}

func (c *Compiler) compileWhile(node *ast.While) {
	cmd := node.Cmd().Name
	c.fprintln("(%s %s", cmd, c.compileExpr(node.Condition))
	c.indent++
	c.compile(node.Body)
	c.trimLineBreak()
	c.writeString(")\n")
	c.indent--
}

func (c *Compiler) compileFor(node *ast.For) {
	cmd := node.Cmd().Name
	left := ""
	if node.Left != nil {
		left = c.compileExpr(node.Left)
	} else {
		ls := make([]string, 0, len(node.List))
		for _, n := range node.List {
			ls = append(ls, c.compileExpr(n))
		}
		rest := ""
		if node.Rest != nil {
			rest = " . " + c.compileExpr(node.Rest)
		}
		left = fmt.Sprintf("(%s%s)", strings.Join(ls, " "), rest)
	}
	right := c.compileExpr(node.Right)
	c.fprintln("(%s %s %s", cmd, left, right)
	c.indent++
	c.compile(node.Body)
	c.trimLineBreak()
	c.writeString(")\n")
	c.indent--
}

func (c *Compiler) compileSingleCmd(node ast.ExCommand) {
	c.fprintln("(%s)", node.Cmd().Name)
}

func (c *Compiler) compileTry(node *ast.Try) {
	c.fprintln("(%s", node.Cmd().Name)
	c.indent++
	c.compile(node.Body)
	for _, n := range node.Catch {
		c.indent--
		if n.Pattern != "" {
			c.fprintln(" %s /%s/", n.Cmd().Name, n.Pattern)
			c.indent++
			c.compile(n.Body)
		} else {
			c.fprintln(" %s", n.Cmd().Name)
			c.indent++
			c.compile(n.Body)
		}
	}
	if node.Finally != nil {
		c.indent--
		c.fprintln(" %s", node.Finally.Cmd().Name)
		c.indent++
		c.compile(node.Finally.Body)
	}
	c.trimLineBreak()
	c.writeString(")\n")
	c.indent--
}

func (c *Compiler) compileThrow(node *ast.Throw) {
	cmd := node.Cmd().Name
	c.fprintln("(%s %s)", cmd, c.compileExpr(node.Expr))
}

func (c *Compiler) compileEchocmd(node *ast.EchoCmd) {
	cmd := node.Cmd().Name
	exprs := make([]string, 0, len(node.Exprs))
	for _, e := range node.Exprs {
		exprs = append(exprs, c.compileExpr(e))
	}
	c.fprintln("(%s %s)", cmd, strings.Join(exprs, " "))
}

func (c *Compiler) compileEchohl(node *ast.Echohl) {
	cmd := node.Cmd().Name
	c.fprintln(`(%s "%s")`, cmd, escape(node.Name, `\"`))
}

func (c *Compiler) compileExecute(node *ast.Execute) {
	cmd := node.Cmd().Name
	list := make([]string, 0, len(node.Exprs))
	for _, e := range node.Exprs {
		list = append(list, c.compileExpr(e))
	}
	c.fprintln("(%s %s)", cmd, strings.Join(list, " "))
}

func (c *Compiler) compileExpr(node ast.Expr) string {
	switch n := node.(type) {
	case *ast.TernaryExpr:
		cond := c.compileExpr(n.Condition)
		l := c.compileExpr(n.Left)
		r := c.compileExpr(n.Right)
		return fmt.Sprintf("(?: %s %s %s)", cond, l, r)
	case *ast.BinaryExpr:
		l := c.compileExpr(n.Left)
		r := c.compileExpr(n.Right)
		op := n.Op.String()
		if op == "." {
			op = "concat"
		}
		return fmt.Sprintf("(%s %s %s)", op, l, r)
	case *ast.UnaryExpr:
		return fmt.Sprintf("(%s %s)", n.Op, c.compileExpr(n.X))
	case *ast.SubscriptExpr:
		l := c.compileExpr(n.Left)
		r := c.compileExpr(n.Right)
		return fmt.Sprintf("(subscript %s %s)", l, r)
	case *ast.SliceExpr:
		x := c.compileExpr(n.X)
		l := "nil"
		if n.Low != nil {
			l = c.compileExpr(n.Low)
		}
		h := "nil"
		if n.High != nil {
			h = c.compileExpr(n.High)
		}
		return fmt.Sprintf("(slice %s %s %s)", x, l, h)
	case *ast.CallExpr:
		name := c.compileExpr(n.Fun)
		if len(n.Args) > 0 {
			args := make([]string, 0, len(n.Args))
			for _, a := range n.Args {
				args = append(args, c.compileExpr(a))
			}
			return fmt.Sprintf("(%s %s)", name, strings.Join(args, " "))
		}
		return fmt.Sprintf("(%s)", name)
	case *ast.DotExpr:
		l := c.compileExpr(n.Left)
		r := c.compileExpr(n.Right)
		return fmt.Sprintf("(dot %s %s)", l, r)
	case *ast.BasicLit:
		return n.Value
	case *ast.List:
		if len(n.Values) == 0 {
			return "(list)"
		}
		vs := make([]string, 0, len(n.Values))
		for _, v := range n.Values {
			vs = append(vs, c.compileExpr(v))
		}
		return fmt.Sprintf("(list %s)", strings.Join(vs, " "))
	case *ast.Dict:
		if len(n.Entries) == 0 {
			return "(dict)"
		}
		kvs := make([]string, 0, len(n.Entries))
		for _, e := range n.Entries {
			kv := fmt.Sprintf("(%s %s)", c.compileExpr(e.Key), c.compileExpr(e.Value))
			kvs = append(kvs, kv)
		}
		return fmt.Sprintf("(dict %s)", strings.Join(kvs, " "))
	case *ast.CurlyName:
		ps := make([]string, 0, len(n.Parts))
		for _, part := range n.Parts {
			ps = append(ps, c.compileExpr(part))
		}
		return strings.Join(ps, "")
	case *ast.CurlyNameLit:
		return n.Value
	case *ast.CurlyNameExpr:
		return fmt.Sprintf("{%s}", c.compileExpr(n.Value))
	case *ast.Ident:
		return n.Name
	case *ast.LambdaExpr:
		params := make([]string, 0, len(n.Params))
		for _, p := range n.Params {
			params = append(params, p.Name)
		}
		return fmt.Sprintf("(lambda (%s) %s)", strings.Join(params, " "), c.compileExpr(n.Expr))
	case *ast.ParenExpr:
		return c.compileExpr(n.X)
	}
	return ""
}

func escape(s string, chars string) string {
	r := ""
	for _, c := range s {
		if strings.IndexRune(chars, c) != -1 {
			r += `\` + string(c)
		} else {
			r += string(c)
		}
	}
	return r
}
