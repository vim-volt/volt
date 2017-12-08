// Package ast provides Vim script AST.
// ref: "go/ast"
package ast

import (
	"github.com/haya14busa/go-vimlparser/token"
)

// Node is the interface for all node types to implement.
type Node interface {
	Pos() Pos // position of first character belonging to the node
}

// Statement is the interface for statement (Ex command or Comment).
type Statement interface {
	Node
	stmtNode()
}

// ExCommand is the interface for Ex-command.
type ExCommand interface {
	Statement
	Cmd() Cmd
}

// Expr is the interface for expression.
type Expr interface {
	Node
	exprNode()
}

// File node represents a Vim script source file.
// Equivalent to NODE_TOPLEVEL of vim-vimlparser.
// vimlparser: TOPLEVEL .body
type File struct {
	Start Pos         // position of start of node.
	Body  []Statement // top-level declarations; or nil
}

func (f *File) Pos() Pos { return f.Start }

// vimlparser: COMMENT .str
type Comment struct {
	Statement
	Quote Pos    // position of `"` starting the comment
	Text  string // comment text (excluding '\n')
}

func (c *Comment) Pos() Pos { return c.Quote }

// vimlparser: EXCMD .ea .str
type Excmd struct {
	Excmd   Pos    // position of starting the excmd
	Command string // Ex comamnd
	ExArg   ExArg  // Ex command arg
}

func (e *Excmd) Pos() Pos { return e.Excmd }
func (e *Excmd) Cmd() Cmd { return *e.ExArg.Cmd }

// vimlparser: FUNCTION .ea .body .left .rlist .attr .endfunction
type Function struct {
	Func        Pos          // position of starting the :function
	ExArg       ExArg        // Ex command arg
	Body        []Statement  // function body
	Name        Expr         // function name
	Params      []*Ident     // parameters
	Attr        FuncAttr     // function attributes
	EndFunction *EndFunction // :endfunction
}

type FuncAttr struct {
	Range   bool
	Abort   bool
	Dict    bool
	Closure bool
}

func (f *Function) Pos() Pos { return f.Func }
func (f *Function) Cmd() Cmd { return *f.ExArg.Cmd }

// vimlparser: ENDFUNCTION .ea
type EndFunction struct {
	EndFunc Pos   // position of starting the :endfunction
	ExArg   ExArg // Ex command arg
}

func (f *EndFunction) Pos() Pos { return f.EndFunc }
func (f *EndFunction) Cmd() Cmd { return *f.ExArg.Cmd }

// vimlparser: DELFUNCTION .ea
type DelFunction struct {
	DelFunc Pos   // position of starting the :delfunction
	ExArg   ExArg // Ex command arg
	Name    Expr  // function name to delete
}

func (f *DelFunction) Pos() Pos { return f.DelFunc }
func (f *DelFunction) Cmd() Cmd { return *f.ExArg.Cmd }

// vimlparser: RETURN .ea .left
type Return struct {
	Return Pos   // position of starting the :return
	ExArg  ExArg // Ex command arg
	Result Expr  // expression to return
}

func (f *Return) Pos() Pos { return f.Return }
func (f *Return) Cmd() Cmd { return *f.ExArg.Cmd }

// vimlparser: EXCALL .ea .left
type ExCall struct {
	ExCall   Pos       // position of starting the :call
	ExArg    ExArg     // Ex command arg
	FuncCall *CallExpr // a function call
}

func (f *ExCall) Pos() Pos { return f.ExCall }
func (f *ExCall) Cmd() Cmd { return *f.ExArg.Cmd }

// vimlparser: LET .ea .op .left .list .rest .right
type Let struct {
	Let   Pos    // position of starting the :let
	ExArg ExArg  // Ex command arg
	Op    string // operator

	// :let {'a'} = 1
	//      ^^^^^ Left
	Left Expr // lhs; or nil

	// :let [{'a'}, b; {'c'}] = [1,2,3,4,5]
	//       ^^^^^^^^  ^^^^^
	//       List      Rest
	List []Expr // lhs list; or nil
	Rest Expr   // rest of lhs list; or nil

	Right Expr // rhs expression.
}

func (f *Let) Pos() Pos { return f.Let }
func (f *Let) Cmd() Cmd { return *f.ExArg.Cmd }

// vimlparser: UNLET .ea .list
type UnLet struct {
	UnLet Pos    // position of starting the :unlet
	ExArg ExArg  // Ex command arg
	List  []Expr // list to unlet
}

func (f *UnLet) Pos() Pos { return f.UnLet }
func (f *UnLet) Cmd() Cmd { return *f.ExArg.Cmd }

// vimlparser: LOCKVAR .ea .depth .list
type LockVar struct {
	LockVar Pos    // position of starting the :lockvar
	ExArg   ExArg  // Ex command arg
	Depth   int    // default: 0
	List    []Expr // list to lockvar
}

func (f *LockVar) Pos() Pos { return f.LockVar }
func (f *LockVar) Cmd() Cmd { return *f.ExArg.Cmd }

// vimlparser: UNLOCKVAR .ea .depth .list
type UnLockVar struct {
	UnLockVar Pos    // position of starting the :lockvar
	ExArg     ExArg  // Ex command arg
	Depth     int    // default: 0
	List      []Expr // list to lockvar
}

func (f *UnLockVar) Pos() Pos { return f.UnLockVar }
func (f *UnLockVar) Cmd() Cmd { return *f.ExArg.Cmd }

// vimlparser: IF .ea .body .cond .elseif .else .endif
type If struct {
	If        Pos         // position of starting the :if
	ExArg     ExArg       // Ex command arg
	Body      []Statement // body of if statement
	Condition Expr        // condition
	ElseIf    []*ElseIf
	Else      *Else
	EndIf     *EndIf
}

func (f *If) Pos() Pos { return f.If }
func (f *If) Cmd() Cmd { return *f.ExArg.Cmd }

// vimlparser: ELSEIF .ea .body .cond
type ElseIf struct {
	ElseIf    Pos         // position of starting the :elseif
	ExArg     ExArg       // Ex command arg
	Body      []Statement // body of elseif statement
	Condition Expr        // condition
}

func (f *ElseIf) Pos() Pos { return f.ElseIf }
func (f *ElseIf) Cmd() Cmd { return *f.ExArg.Cmd }

// vimlparser: ELSE .ea .body
type Else struct {
	Else  Pos         // position of starting the :else
	ExArg ExArg       // Ex command arg
	Body  []Statement // body of else statement
}

func (f *Else) Pos() Pos { return f.Else }
func (f *Else) Cmd() Cmd { return *f.ExArg.Cmd }

// vimlparser: ENDIF .ea
type EndIf struct {
	EndIf Pos   // position of starting the :endif
	ExArg ExArg // Ex command arg
}

func (f *EndIf) Pos() Pos { return f.EndIf }
func (f *EndIf) Cmd() Cmd { return *f.ExArg.Cmd }

// vimlparser: WHILE .ea .body .cond .endwhile
type While struct {
	While     Pos         // position of starting the :while
	ExArg     ExArg       // Ex command arg
	Body      []Statement // body of while statement
	Condition Expr        // condition
	EndWhile  *EndWhile
}

func (f *While) Pos() Pos { return f.While }
func (f *While) Cmd() Cmd { return *f.ExArg.Cmd }

// vimlparser: ENDWHILE .ea
type EndWhile struct {
	EndWhile Pos   // position of starting the :endwhile
	ExArg    ExArg // Ex command arg
}

func (f *EndWhile) Pos() Pos { return f.EndWhile }
func (f *EndWhile) Cmd() Cmd { return *f.ExArg.Cmd }

// vimlparser: FOR .ea .body .left .list .rest .right .endfor
type For struct {
	For   Pos         // position of starting the :for
	ExArg ExArg       // Ex command arg
	Body  []Statement // body of for statement

	// :for {'a'} in right
	//      ^^^^^ Left
	Left Expr // lhs; or nil

	// :for [{'a'}, b; {'c'}] in right
	//       ^^^^^^^^  ^^^^^
	//       List      Rest
	List []Expr // lhs list; or nil
	Rest Expr   // rest of lhs list; or nil

	Right  Expr // rhs expression.
	EndFor *EndFor
}

func (f *For) Pos() Pos { return f.For }
func (f *For) Cmd() Cmd { return *f.ExArg.Cmd }

// vimlparser: ENDFOR .ea
type EndFor struct {
	EndFor Pos   // position of starting the :endfor
	ExArg  ExArg // Ex command arg
}

func (f *EndFor) Pos() Pos { return f.EndFor }
func (f *EndFor) Cmd() Cmd { return *f.ExArg.Cmd }

// vimlparser: CONTINUE .ea
type Continue struct {
	Continue Pos   // position of starting the :continue
	ExArg    ExArg // Ex command arg
}

func (f *Continue) Pos() Pos { return f.Continue }
func (f *Continue) Cmd() Cmd { return *f.ExArg.Cmd }

// vimlparser: BREAK .ea
type Break struct {
	Break Pos   // position of starting the :break
	ExArg ExArg // Ex command arg
}

func (f *Break) Pos() Pos { return f.Break }
func (f *Break) Cmd() Cmd { return *f.ExArg.Cmd }

// vimlparser: TRY .ea .body .catch .finally .endtry
type Try struct {
	Try     Pos         // position of starting the :try
	ExArg   ExArg       // Ex command arg
	Body    []Statement // body of try statement
	Catch   []*Catch
	Finally *Finally
	EndTry  *EndTry
}

func (f *Try) Pos() Pos { return f.Try }
func (f *Try) Cmd() Cmd { return *f.ExArg.Cmd }

// vimlparser: CATCH .ea .body .pattern
type Catch struct {
	Catch   Pos         // position of starting the :catch
	ExArg   ExArg       // Ex command arg
	Body    []Statement // body of catch statement
	Pattern string      // pattern
}

func (f *Catch) Pos() Pos { return f.Catch }
func (f *Catch) Cmd() Cmd { return *f.ExArg.Cmd }

// vimlparser: FINALLY .ea .body
type Finally struct {
	Finally Pos         // position of starting the :finally
	ExArg   ExArg       // Ex command arg
	Body    []Statement // body of else statement
}

func (f *Finally) Pos() Pos { return f.Finally }
func (f *Finally) Cmd() Cmd { return *f.ExArg.Cmd }

// vimlparser: ENDTRY .ea
type EndTry struct {
	EndTry Pos   // position of starting the :endtry
	ExArg  ExArg // Ex command arg
}

func (f *EndTry) Pos() Pos { return f.EndTry }
func (f *EndTry) Cmd() Cmd { return *f.ExArg.Cmd }

// vimlparser: THROW .ea .left
// :throw {Expr}
type Throw struct {
	Throw Pos   // position of starting the :throw
	ExArg ExArg // Ex command arg
	Expr  Expr
}

func (f *Throw) Pos() Pos { return f.Throw }
func (f *Throw) Cmd() Cmd { return *f.ExArg.Cmd }

// vimlparser: ECHO .ea .list
// vimlparser: ECHON .ea .list
// vimlparser: ECHOMSG .ea .list
// vimlparser: ECHOERR .ea .list
// :{echocmd} {Expr}..
// {echocmd}: echo, echon, echomsg, echoerr
type EchoCmd struct {
	Start   Pos    // position of starting the echo-command
	CmdName string // echo-command name
	ExArg   ExArg  // Ex command arg
	Exprs   []Expr
}

func (f *EchoCmd) Pos() Pos { return f.Start }
func (f *EchoCmd) Cmd() Cmd { return *f.ExArg.Cmd }

// vimlparser: ECHOHL .ea .str
// :echohl {name}
type Echohl struct {
	Echohl Pos   // position of starting the :echohl
	ExArg  ExArg // Ex command arg
	Name   string
}

func (f *Echohl) Pos() Pos { return f.Echohl }
func (f *Echohl) Cmd() Cmd { return *f.ExArg.Cmd }

// vimlparser: EXECUTE .ea .list
// :execute {Expr}..
type Execute struct {
	Execute Pos   // position of starting the :execute
	ExArg   ExArg // Ex command arg
	Exprs   []Expr
}

func (f *Execute) Pos() Pos { return f.Execute }
func (f *Execute) Cmd() Cmd { return *f.ExArg.Cmd }

// vimlparser: TERNARY .cond .left .right
// Condition ? Left : Right
type TernaryExpr struct {
	Ternary   Pos // position of starting the :execute
	Condition Expr
	Left      Expr
	Right     Expr
}

func (f *TernaryExpr) Pos() Pos { return f.Ternary }

type BinaryExpr struct {
	Left  Expr        // left operand
	OpPos Pos         // position of Op
	Op    token.Token // operator
	Right Expr        // right operand
}

func (f *BinaryExpr) Pos() Pos { return f.OpPos }

type UnaryExpr struct {
	OpPos Pos         // position of Op
	Op    token.Token // operator
	X     Expr        // operand
}

func (f *UnaryExpr) Pos() Pos { return f.OpPos }

// Left[Right]
type SubscriptExpr struct {
	Lbrack Pos // position of "["
	Left   Expr
	Right  Expr
}

func (f *SubscriptExpr) Pos() Pos { return f.Lbrack }

// X[Low:High]
type SliceExpr struct {
	X      Expr // expression
	Lbrack Pos  // position of "["
	Low    Expr // begin of slice range; or nil
	High   Expr // end of slice range; or nil
}

func (f *SliceExpr) Pos() Pos { return f.Lbrack }

// vimlparser: CALL .left .rlist
type CallExpr struct {
	Fun    Expr   // function expression
	Lparen Pos    // position of "("
	Args   []Expr // function arguments; or nil
}

func (c *CallExpr) Pos() Pos { return c.Lparen }

// Left.Right
// vimlparser: Dot .left .right
type DotExpr struct {
	Left  Expr
	Dot   Pos // position of "."
	Right *Ident
}

func (c *DotExpr) Pos() Pos { return c.Dot }

type BasicLit struct {
	ValuePos Pos         // literal position
	Kind     token.Token // token.INT, token.STRING, token.OPTION, token.ENV, token.REG
	Value    string
}

func (c *BasicLit) Pos() Pos { return c.ValuePos }

type List struct {
	Lsquare Pos // position of "["
	Values  []Expr
}

func (c *List) Pos() Pos { return c.Lsquare }

type Dict struct {
	Lcurlybrace Pos // position of "{"
	Entries     []KeyValue
}

func (c *Dict) Pos() Pos { return c.Lcurlybrace }

type KeyValue struct {
	Key   Expr
	Value Expr
	// TODO: want Pos data...
}

// aaa{x{y{1+2}}}bbb
// ^^^^^^^^^^^^^^^^^ <- CurlyName
type CurlyName struct {
	CurlyName Pos // position
	Parts     []CurlyNamePart
}

func (c *CurlyName) Pos() Pos { return c.CurlyName }

type CurlyNamePart interface {
	Expr
	IsCurlyExpr() bool
}

// aaa{x{y{1+2}}}bbb
// ^^^           ^^^ <- CurlyNameLit
type CurlyNameLit struct {
	CurlyNameLit Pos // position
	Value        string
}

func (c *CurlyNameLit) Pos() Pos          { return c.CurlyNameLit }
func (c *CurlyNameLit) IsCurlyExpr() bool { return false }

// aaa{x{y{1+2}}}bbb
//    ^^^^^^^^^^^    <- CurlyNameExpr
type CurlyNameExpr struct {
	CurlyNameExpr Pos // position
	Value         Expr
}

func (c *CurlyNameExpr) Pos() Pos          { return c.CurlyNameExpr }
func (c *CurlyNameExpr) IsCurlyExpr() bool { return true }

// An Ident node represents an identifier.
type Ident struct {
	NamePos Pos    // identifier position
	Name    string // identifier name
}

func (i *Ident) Pos() Pos { return i.NamePos }

// LambdaExpr node represents lambda.
// vimlparsr: LAMBDA .rlist .left
// { Params -> Expr }
type LambdaExpr struct {
	Lcurlybrace Pos      // position of "{"
	Params      []*Ident // parameters
	Expr        Expr
}

func (i *LambdaExpr) Pos() Pos { return i.Lcurlybrace }

// ParenExpr node represents a parenthesized expression.
// vimlparsr: PARENEXPR .value
type ParenExpr struct {
	Lparen Pos  // position of "("
	X      Expr // parenthesized expression
}

func (i *ParenExpr) Pos() Pos { return i.Lparen }

// stmtNode() ensures that only ExComamnd and Comment nodes can be assigned to
// an Statement.
//
func (*Break) stmtNode()      {}
func (*Catch) stmtNode()      {}
func (*Continue) stmtNode()   {}
func (DelFunction) stmtNode() {}
func (*EchoCmd) stmtNode()    {}
func (*Echohl) stmtNode()     {}
func (*Else) stmtNode()       {}
func (*ElseIf) stmtNode()     {}
func (*EndFor) stmtNode()     {}
func (EndFunction) stmtNode() {}
func (*EndIf) stmtNode()      {}
func (*EndTry) stmtNode()     {}
func (*EndWhile) stmtNode()   {}
func (*ExCall) stmtNode()     {}
func (Excmd) stmtNode()       {}
func (*Execute) stmtNode()    {}
func (*Finally) stmtNode()    {}
func (*For) stmtNode()        {}
func (Function) stmtNode()    {}
func (*If) stmtNode()         {}
func (*Let) stmtNode()        {}
func (*LockVar) stmtNode()    {}
func (*Return) stmtNode()     {}
func (*Throw) stmtNode()      {}
func (*Try) stmtNode()        {}
func (*UnLet) stmtNode()      {}
func (*UnLockVar) stmtNode()  {}
func (*While) stmtNode()      {}

func (*Comment) stmtNode() {}

// exprNode() ensures that only expression nodes can be assigned to an Expr.
//
func (*TernaryExpr) exprNode()   {}
func (*BinaryExpr) exprNode()    {}
func (*UnaryExpr) exprNode()     {}
func (*SubscriptExpr) exprNode() {}
func (*SliceExpr) exprNode()     {}
func (*CallExpr) exprNode()      {}
func (*DotExpr) exprNode()       {}
func (*BasicLit) exprNode()      {}
func (*List) exprNode()          {}
func (*Dict) exprNode()          {}
func (*CurlyName) exprNode()     {}
func (*CurlyNameLit) exprNode()  {}
func (*CurlyNameExpr) exprNode() {}
func (*Ident) exprNode()         {}
func (*LambdaExpr) exprNode()    {}
func (*ParenExpr) exprNode()     {}
