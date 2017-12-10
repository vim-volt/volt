package vimlparser

import (
	"fmt"

	"github.com/haya14busa/go-vimlparser/ast"
	"github.com/haya14busa/go-vimlparser/token"
)

// Parse parses Vim script in reader and returns Node.
func (p *VimLParser) Parse(reader *StringReader, filename string) ast.Node {
	return newAstNode(p.parse(reader), filename)
}

// Parse parses Vim script expression.
func (p *ExprParser) Parse() ast.Expr {
	return newExprNode(p.parse(), "")
}

// ----

// newAstNode converts internal node type to ast.Node.
// n.type_ must no be zero value.
// n.pos must no be nil except TOPLEVEL node.
func newAstNode(n *VimNode, filename string) ast.Node {
	if n == nil {
		return nil
	}

	// TOPLEVEL doens't have position...?
	var pos ast.Pos
	if p := newPos(n.pos, filename); p != nil {
		pos = *p
	} else {
		pos = ast.Pos{Offset: 0, Line: 1, Column: 1, Filename: filename}
	}

	switch n.type_ {

	case NODE_TOPLEVEL:
		return &ast.File{Start: pos, Body: newBody(*n, filename)}

	case NODE_COMMENT:
		return &ast.Comment{
			Quote: pos,
			Text:  n.str,
		}

	case NODE_EXCMD:
		return &ast.Excmd{
			Excmd:   pos,
			ExArg:   newExArg(*n.ea, filename),
			Command: n.str,
		}

	case NODE_FUNCTION:
		attr := ast.FuncAttr{}
		if n.attr != nil {
			attr = ast.FuncAttr{
				Range:   n.attr.range_,
				Abort:   n.attr.abort,
				Dict:    n.attr.dict,
				Closure: n.attr.closure,
			}
		}
		return &ast.Function{
			Func:        pos,
			ExArg:       newExArg(*n.ea, filename),
			Body:        newBody(*n, filename),
			Name:        newExprNode(n.left, filename),
			Params:      newIdents(*n, filename),
			Attr:        attr,
			EndFunction: newAstNode(n.endfunction, filename).(*ast.EndFunction),
		}

	case NODE_ENDFUNCTION:
		return &ast.EndFunction{
			EndFunc: pos,
			ExArg:   newExArg(*n.ea, filename),
		}

	case NODE_DELFUNCTION:
		return &ast.DelFunction{
			DelFunc: pos,
			ExArg:   newExArg(*n.ea, filename),
			Name:    newExprNode(n.left, filename),
		}

	case NODE_RETURN:
		return &ast.Return{
			Return: pos,
			ExArg:  newExArg(*n.ea, filename),
			Result: newExprNode(n.left, filename),
		}

	case NODE_EXCALL:
		return &ast.ExCall{
			ExCall:   pos,
			ExArg:    newExArg(*n.ea, filename),
			FuncCall: newAstNode(n.left, filename).(*ast.CallExpr),
		}

	case NODE_LET:
		return &ast.Let{
			Let:   pos,
			ExArg: newExArg(*n.ea, filename),
			Op:    n.op,
			Left:  newExprNode(n.left, filename),
			List:  newList(*n, filename),
			Rest:  newExprNode(n.rest, filename),
			Right: newExprNode(n.right, filename),
		}

	case NODE_UNLET:
		return &ast.UnLet{
			UnLet: pos,
			ExArg: newExArg(*n.ea, filename),
			List:  newList(*n, filename),
		}

	case NODE_LOCKVAR:
		return &ast.LockVar{
			LockVar: pos,
			ExArg:   newExArg(*n.ea, filename),
			Depth:   n.depth,
			List:    newList(*n, filename),
		}

	case NODE_UNLOCKVAR:
		return &ast.UnLockVar{
			UnLockVar: pos,
			ExArg:     newExArg(*n.ea, filename),
			Depth:     n.depth,
			List:      newList(*n, filename),
		}

	case NODE_IF:
		var elifs []*ast.ElseIf
		if n.elseif != nil {
			elifs = make([]*ast.ElseIf, 0, len(n.elseif))
		}
		for _, node := range n.elseif {
			if node != nil { // conservative
				elifs = append(elifs, newAstNode(node, filename).(*ast.ElseIf))
			}
		}
		var els *ast.Else
		if n.else_ != nil {
			els = newAstNode(n.else_, filename).(*ast.Else)
		}
		return &ast.If{
			If:        pos,
			ExArg:     newExArg(*n.ea, filename),
			Body:      newBody(*n, filename),
			Condition: newExprNode(n.cond, filename),
			ElseIf:    elifs,
			Else:      els,
			EndIf:     newAstNode(n.endif, filename).(*ast.EndIf),
		}

	case NODE_ELSEIF:
		return &ast.ElseIf{
			ElseIf:    pos,
			ExArg:     newExArg(*n.ea, filename),
			Body:      newBody(*n, filename),
			Condition: newExprNode(n.cond, filename),
		}

	case NODE_ELSE:
		return &ast.Else{
			Else:  pos,
			ExArg: newExArg(*n.ea, filename),
			Body:  newBody(*n, filename),
		}

	case NODE_ENDIF:
		return &ast.EndIf{
			EndIf: pos,
			ExArg: newExArg(*n.ea, filename),
		}

	case NODE_WHILE:
		return &ast.While{
			While:     pos,
			ExArg:     newExArg(*n.ea, filename),
			Body:      newBody(*n, filename),
			Condition: newExprNode(n.cond, filename),
			EndWhile:  newAstNode(n.endwhile, filename).(*ast.EndWhile),
		}

	case NODE_ENDWHILE:
		return &ast.EndWhile{
			EndWhile: pos,
			ExArg:    newExArg(*n.ea, filename),
		}

	case NODE_FOR:
		return &ast.For{
			For:    pos,
			ExArg:  newExArg(*n.ea, filename),
			Body:   newBody(*n, filename),
			Left:   newExprNode(n.left, filename),
			List:   newList(*n, filename),
			Rest:   newExprNode(n.rest, filename),
			Right:  newExprNode(n.right, filename),
			EndFor: newAstNode(n.endfor, filename).(*ast.EndFor),
		}

	case NODE_ENDFOR:
		return &ast.EndFor{
			EndFor: pos,
			ExArg:  newExArg(*n.ea, filename),
		}

	case NODE_CONTINUE:
		return &ast.Continue{
			Continue: pos,
			ExArg:    newExArg(*n.ea, filename),
		}

	case NODE_BREAK:
		return &ast.Break{
			Break: pos,
			ExArg: newExArg(*n.ea, filename),
		}

	case NODE_TRY:
		var catches []*ast.Catch
		if n.catch != nil {
			catches = make([]*ast.Catch, 0, len(n.catch))
		}
		for _, node := range n.catch {
			if node != nil { // conservative
				catches = append(catches, newAstNode(node, filename).(*ast.Catch))
			}
		}
		var finally *ast.Finally
		if n.finally != nil {
			finally = newAstNode(n.finally, filename).(*ast.Finally)
		}
		return &ast.Try{
			Try:     pos,
			ExArg:   newExArg(*n.ea, filename),
			Body:    newBody(*n, filename),
			Catch:   catches,
			Finally: finally,
			EndTry:  newAstNode(n.endtry, filename).(*ast.EndTry),
		}

	case NODE_CATCH:
		return &ast.Catch{
			Catch:   pos,
			ExArg:   newExArg(*n.ea, filename),
			Body:    newBody(*n, filename),
			Pattern: n.pattern,
		}

	case NODE_FINALLY:
		return &ast.Finally{
			Finally: pos,
			ExArg:   newExArg(*n.ea, filename),
			Body:    newBody(*n, filename),
		}

	case NODE_ENDTRY:
		return &ast.EndTry{
			EndTry: pos,
			ExArg:  newExArg(*n.ea, filename),
		}

	case NODE_THROW:
		return &ast.Throw{
			Throw: pos,
			ExArg: newExArg(*n.ea, filename),
			Expr:  newExprNode(n.left, filename),
		}

	case NODE_ECHO, NODE_ECHON, NODE_ECHOMSG, NODE_ECHOERR:
		return &ast.EchoCmd{
			Start:   pos,
			CmdName: n.ea.cmd.name,
			ExArg:   newExArg(*n.ea, filename),
			Exprs:   newList(*n, filename),
		}

	case NODE_ECHOHL:
		return &ast.Echohl{
			Echohl: pos,
			ExArg:  newExArg(*n.ea, filename),
			Name:   n.str,
		}

	case NODE_EXECUTE:
		return &ast.Execute{
			Execute: pos,
			ExArg:   newExArg(*n.ea, filename),
			Exprs:   newList(*n, filename),
		}

	case NODE_TERNARY:
		return &ast.TernaryExpr{
			Ternary:   pos,
			Condition: newExprNode(n.cond, filename),
			Left:      newExprNode(n.left, filename),
			Right:     newExprNode(n.right, filename),
		}

	case NODE_OR, NODE_AND, NODE_EQUAL, NODE_EQUALCI, NODE_EQUALCS,
		NODE_NEQUAL, NODE_NEQUALCI, NODE_NEQUALCS, NODE_GREATER,
		NODE_GREATERCI, NODE_GREATERCS, NODE_GEQUAL, NODE_GEQUALCI,
		NODE_GEQUALCS, NODE_SMALLER, NODE_SMALLERCI, NODE_SMALLERCS,
		NODE_SEQUAL, NODE_SEQUALCI, NODE_SEQUALCS, NODE_MATCH,
		NODE_MATCHCI, NODE_MATCHCS, NODE_NOMATCH, NODE_NOMATCHCI,
		NODE_NOMATCHCS, NODE_IS, NODE_ISCI, NODE_ISCS, NODE_ISNOT,
		NODE_ISNOTCI, NODE_ISNOTCS, NODE_ADD, NODE_SUBTRACT, NODE_CONCAT,
		NODE_MULTIPLY, NODE_DIVIDE, NODE_REMAINDER:
		return &ast.BinaryExpr{
			Left:  newExprNode(n.left, filename),
			OpPos: pos,
			Op:    opToken(n.type_),
			Right: newExprNode(n.right, filename),
		}

	case NODE_NOT, NODE_MINUS, NODE_PLUS:
		return &ast.UnaryExpr{
			OpPos: pos,
			Op:    opToken(n.type_),
			X:     newExprNode(n.left, filename),
		}

	case NODE_SUBSCRIPT:
		return &ast.SubscriptExpr{
			Lbrack: pos,
			Left:   newExprNode(n.left, filename),
			Right:  newExprNode(n.right, filename),
		}

	case NODE_SLICE:
		return &ast.SliceExpr{
			Lbrack: pos,
			X:      newExprNode(n.left, filename),
			Low:    newExprNode(n.rlist[0], filename),
			High:   newExprNode(n.rlist[1], filename),
		}

	case NODE_CALL:
		return &ast.CallExpr{
			Lparen: pos,
			Fun:    newExprNode(n.left, filename),
			Args:   newRlist(*n, filename),
		}

	case NODE_DOT:
		return &ast.DotExpr{
			Left:  newExprNode(n.left, filename),
			Dot:   pos,
			Right: newAstNode(n.right, filename).(*ast.Ident),
		}

	case NODE_NUMBER:
		return &ast.BasicLit{
			ValuePos: pos,
			Kind:     token.NUMBER,
			Value:    n.value.(string),
		}
	case NODE_STRING:
		return &ast.BasicLit{
			ValuePos: pos,
			Kind:     token.STRING,
			Value:    n.value.(string),
		}
	case NODE_LIST:
		return &ast.List{
			Lsquare: pos,
			Values:  newValues(*n, filename),
		}

	case NODE_DICT:
		entries := n.value.([]interface{})
		kvs := make([]ast.KeyValue, 0, len(entries))
		for _, nn := range entries {
			kv := nn.([]interface{})
			k := newExprNode(kv[0].(*VimNode), filename)
			v := newExprNode(kv[1].(*VimNode), filename)
			kvs = append(kvs, ast.KeyValue{Key: k, Value: v})
		}
		return &ast.Dict{
			Lcurlybrace: pos,
			Entries:     kvs,
		}

	case NODE_OPTION:
		return &ast.BasicLit{
			ValuePos: pos,
			Kind:     token.OPTION,
			Value:    n.value.(string),
		}
	case NODE_IDENTIFIER:
		return &ast.Ident{
			NamePos: pos,
			Name:    n.value.(string),
		}

	case NODE_CURLYNAME:
		var parts []ast.CurlyNamePart
		for _, n := range n.value.([]*VimNode) {
			node := newAstNode(n, filename)
			parts = append(parts, node.(ast.CurlyNamePart))
		}
		return &ast.CurlyName{
			CurlyName: pos,
			Parts:     parts,
		}

	case NODE_ENV:
		return &ast.BasicLit{
			ValuePos: pos,
			Kind:     token.ENV,
			Value:    n.value.(string),
		}

	case NODE_REG:
		return &ast.BasicLit{
			ValuePos: pos,
			Kind:     token.REG,
			Value:    n.value.(string),
		}

	case NODE_CURLYNAMEPART:
		return &ast.CurlyNameLit{
			CurlyNameLit: pos,
			Value:        n.value.(string),
		}

	case NODE_CURLYNAMEEXPR:
		n := n.value.(*VimNode)
		return &ast.CurlyNameExpr{
			CurlyNameExpr: pos,
			Value:         newExprNode(n, filename),
		}

	case NODE_LAMBDA:
		return &ast.LambdaExpr{
			Lcurlybrace: pos,
			Params:      newIdents(*n, filename),
			Expr:        newExprNode(n.left, filename),
		}

	case NODE_PARENEXPR:
		n := n.value.(*VimNode)
		return &ast.ParenExpr{
			Lparen: pos,
			X:      newExprNode(n, filename),
		}

	}
	panic(fmt.Errorf("Unknown node type: %v, node: %v", n.type_, n))
}

func newExprNode(n *VimNode, filename string) ast.Expr {
	node, _ := newAstNode(n, filename).(ast.Expr)
	return node
}

func newPos(p *pos, filename string) *ast.Pos {
	if p == nil {
		return nil
	}
	return &ast.Pos{
		Offset:   p.offset,
		Line:     p.lnum,
		Column:   p.col,
		Filename: filename,
	}
}

func newExArg(ea ExArg, filename string) ast.ExArg {
	return ast.ExArg{
		Forceit:    ea.forceit,
		AddrCount:  ea.addr_count,
		Line1:      ea.line1,
		Line2:      ea.line2,
		Flags:      ea.flags,
		DoEcmdCmd:  ea.do_ecmd_cmd,
		DoEcmdLnum: ea.do_ecmd_lnum,
		Append:     ea.append,
		Usefilter:  ea.usefilter,
		Amount:     ea.amount,
		Regname:    ea.regname,
		ForceBin:   ea.force_bin,
		ReadEdit:   ea.read_edit,
		ForceFf:    ea.force_ff,
		ForceEnc:   ea.force_enc,
		BadChar:    ea.bad_char,
		Linepos:    newPos(ea.linepos, filename),
		Cmdpos:     newPos(ea.cmdpos, filename),
		Argpos:     newPos(ea.argpos, filename),
		Cmd:        newCmd(ea.cmd),
		Modifiers:  ea.modifiers,
		Range:      ea.range_,
		Argopt:     ea.argopt,
		Argcmd:     ea.argcmd,
	}
}

func newCmd(c *Cmd) *ast.Cmd {
	if c == nil {
		return nil
	}
	return &ast.Cmd{
		Name:   c.name,
		Minlen: c.minlen,
		Flags:  c.flags,
		Parser: c.parser,
	}
}

func newBody(n VimNode, filename string) []ast.Statement {
	var body []ast.Statement
	if n.body != nil {
		body = make([]ast.Statement, 0, len(n.body))
	}
	for _, node := range n.body {
		if node != nil { // conservative
			body = append(body, newAstNode(node, filename).(ast.Statement))
		}
	}
	return body
}

func newIdents(n VimNode, filename string) []*ast.Ident {
	var idents []*ast.Ident
	if n.rlist != nil {
		idents = make([]*ast.Ident, 0, len(n.rlist))
	}
	for _, node := range n.rlist {
		if node != nil { // conservative
			idents = append(idents, newAstNode(node, filename).(*ast.Ident))
		}
	}
	return idents
}

func newRlist(n VimNode, filename string) []ast.Expr {
	var exprs []ast.Expr
	if n.rlist != nil {
		exprs = make([]ast.Expr, 0, len(n.rlist))
	}
	for _, node := range n.rlist {
		if node != nil { // conservative
			exprs = append(exprs, newExprNode(node, filename))
		}
	}
	return exprs
}

func newList(n VimNode, filename string) []ast.Expr {
	var list []ast.Expr
	if n.list != nil {
		list = make([]ast.Expr, 0, len(n.list))
	}
	for _, node := range n.list {
		if node != nil { // conservative
			list = append(list, newExprNode(node, filename))
		}
	}
	return list
}

func newValues(n VimNode, filename string) []ast.Expr {
	var values []ast.Expr
	for _, v := range n.value.([]interface{}) {
		n := v.(*VimNode)
		values = append(values, newExprNode(n, filename))
	}
	return values
}

func opToken(nodeType int) token.Token {
	switch nodeType {
	case NODE_OR:
		return token.OROR
	case NODE_AND:
		return token.ANDAND
	case NODE_EQUAL:
		return token.EQEQ
	case NODE_EQUALCI:
		return token.EQEQCI
	case NODE_EQUALCS:
		return token.EQEQCS
	case NODE_NEQUAL:
		return token.NEQ
	case NODE_NEQUALCI:
		return token.NEQCI
	case NODE_NEQUALCS:
		return token.NEQCS
	case NODE_GREATER:
		return token.GT
	case NODE_GREATERCI:
		return token.GTCI
	case NODE_GREATERCS:
		return token.GTCS
	case NODE_GEQUAL:
		return token.GTEQ
	case NODE_GEQUALCI:
		return token.GTEQCI
	case NODE_GEQUALCS:
		return token.GTEQCS
	case NODE_SMALLER:
		return token.LT
	case NODE_SMALLERCI:
		return token.LTCI
	case NODE_SMALLERCS:
		return token.LTCS
	case NODE_SEQUAL:
		return token.LTEQ
	case NODE_SEQUALCI:
		return token.LTEQCI
	case NODE_SEQUALCS:
		return token.LTEQCS
	case NODE_MATCH:
		return token.MATCH
	case NODE_MATCHCI:
		return token.MATCHCI
	case NODE_MATCHCS:
		return token.MATCHCS
	case NODE_NOMATCH:
		return token.NOMATCH
	case NODE_NOMATCHCI:
		return token.NOMATCHCI
	case NODE_NOMATCHCS:
		return token.NOMATCHCS
	case NODE_IS:
		return token.IS
	case NODE_ISCI:
		return token.ISCI
	case NODE_ISCS:
		return token.ISCS
	case NODE_ISNOT:
		return token.ISNOT
	case NODE_ISNOTCI:
		return token.ISNOTCI
	case NODE_ISNOTCS:
		return token.ISNOTCS
	case NODE_ADD:
		return token.PLUS
	case NODE_SUBTRACT:
		return token.MINUS
	case NODE_CONCAT:
		return token.DOT
	case NODE_MULTIPLY:
		return token.STAR
	case NODE_DIVIDE:
		return token.SLASH
	case NODE_REMAINDER:
		return token.PERCENT
	case NODE_NOT:
		return token.NOT
	case NODE_MINUS:
		return token.MINUS
	case NODE_PLUS:
		return token.PLUS
	}
	return token.ILLEGAL
}
