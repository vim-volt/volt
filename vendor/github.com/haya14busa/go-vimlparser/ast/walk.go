package ast

import "fmt"

// A Visitor's Visit method is invoked for each node encountered by Walk.
// If the result visitor w is not nil, Walk visits each of the children
// of node with the visitor w, followed by a call of w.Visit(nil).
// ref: https://golang.org/pkg/go/ast/#Visitor
type Visitor interface {
	Visit(node Node) (w Visitor)
}

// Walk traverses an AST in depth-first order: It starts by calling
// v.Visit(node); node must not be nil. If the visitor w returned by
// v.Visit(node) is not nil, Walk is invoked recursively with visitor
// w for each of the non-nil children of node, followed by a call of
// w.Visit(nil).
//
func Walk(v Visitor, node Node) {
	if node == nil {
		return
	}

	if v = v.Visit(node); v == nil {
		return
	}

	// walk children
	// (the order of the cases matches the order
	// of the corresponding node types in newAstNode func
	// in github.com/haya14busa/go-vimlparser/go/export.go)
	switch n := node.(type) {
	case *File:
		walkStmtList(v, n.Body)

	case *Comment: // nothing to do

	case *Excmd: // nothing to do

	case *Function:
		Walk(v, n.Name)
		walkIdentList(v, n.Params)
		walkStmtList(v, n.Body)
		Walk(v, n.EndFunction)

	case *EndFunction: // nothing to do

	case *DelFunction:
		Walk(v, n.Name)

	case *Return:
		Walk(v, n.Result)

	case *ExCall:
		Walk(v, n.FuncCall)

	case *Let:
		Walk(v, n.Left)
		walkExprList(v, n.List)
		Walk(v, n.Rest)
		Walk(v, n.Right)

	case *UnLet:
		walkExprList(v, n.List)

	case *LockVar:
		walkExprList(v, n.List)

	case *UnLockVar:
		walkExprList(v, n.List)

	case *If:
		Walk(v, n.Condition)
		walkStmtList(v, n.Body)
		for _, elseif := range n.ElseIf {
			Walk(v, elseif)
		}
		if n.Else != nil {
			Walk(v, n.Else)
		}
		if n.EndIf != nil {
			Walk(v, n.EndIf)
		}

	case *ElseIf:
		Walk(v, n.Condition)
		walkStmtList(v, n.Body)

	case *Else:
		walkStmtList(v, n.Body)

	case *EndIf: // nothing to do

	case *While:
		Walk(v, n.Condition)
		walkStmtList(v, n.Body)
		Walk(v, n.EndWhile)

	case *EndWhile: // nothing to do

	case *For:
		Walk(v, n.Left)
		walkExprList(v, n.List)
		Walk(v, n.Rest)
		Walk(v, n.Right)
		walkStmtList(v, n.Body)
		Walk(v, n.EndFor)

	case *EndFor: // nothing to do

	case *Continue: // nothing to do

	case *Break: // nothing to do

	case *Try:
		walkStmtList(v, n.Body)
		for _, c := range n.Catch {
			Walk(v, c)
		}
		if n.Finally != nil {
			Walk(v, n.Finally)
		}
		if n.EndTry != nil {
			Walk(v, n.EndTry)
		}

	case *Catch:
		walkStmtList(v, n.Body)

	case *Finally:
		walkStmtList(v, n.Body)

	case *EndTry: // nothing to do

	case *Throw:
		Walk(v, n.Expr)

	case *EchoCmd:
		walkExprList(v, n.Exprs)

	case *Echohl: // nothing to do

	case *Execute:
		walkExprList(v, n.Exprs)

	case *TernaryExpr:
		Walk(v, n.Condition)
		Walk(v, n.Left)
		Walk(v, n.Right)

	case *BinaryExpr:
		Walk(v, n.Left)
		Walk(v, n.Right)

	case *UnaryExpr:
		Walk(v, n.X)

	case *SubscriptExpr:
		Walk(v, n.Left)
		Walk(v, n.Right)

	case *SliceExpr:
		Walk(v, n.X)
		Walk(v, n.Low)
		Walk(v, n.High)

	case *CallExpr:
		Walk(v, n.Fun)
		walkExprList(v, n.Args)

	case *DotExpr:
		Walk(v, n.Left)
		Walk(v, n.Right)

	case *BasicLit: // nothing to do

	case *List:
		walkExprList(v, n.Values)

	case *Dict:
		for _, e := range n.Entries {
			Walk(v, e.Key)
			Walk(v, e.Value)
		}

	case *Ident: // nothing to do

	case *CurlyName:
		for _, c := range n.Parts {
			Walk(v, c)
		}

	case *CurlyNameLit: // nothing to do

	case *CurlyNameExpr:
		Walk(v, n.Value)

	case *LambdaExpr:
		walkIdentList(v, n.Params)

	case *ParenExpr:
		Walk(v, n.X)

	default:
		panic(fmt.Sprintf("ast.Walk: unexpected node type %T", n))
	}
	v.Visit(nil)
}

// Helper functions for common node lists. They may be empty.

func walkIdentList(v Visitor, list []*Ident) {
	for _, x := range list {
		Walk(v, x)
	}
}

func walkExprList(v Visitor, list []Expr) {
	for _, x := range list {
		Walk(v, x)
	}
}

func walkStmtList(v Visitor, list []Statement) {
	for _, x := range list {
		Walk(v, x)
	}
}

type inspector func(Node) bool

func (f inspector) Visit(node Node) Visitor {
	if node != nil && f(node) {
		return f
	}
	return nil
}

// Inspect traverses an AST in depth-first order: It starts by calling
// f(node); node must not be nil. If f returns true, Inspect invokes f
// recursively for each of the non-nil children of node, followed by a
// call of f(nil).
//
func Inspect(node Node, f func(Node) bool) {
	Walk(inspector(f), node)
}
