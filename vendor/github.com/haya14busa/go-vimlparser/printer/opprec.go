package printer

import (
	"fmt"

	"github.com/haya14busa/go-vimlparser/ast"
	"github.com/haya14busa/go-vimlparser/token"
)

// opprec returns operator precedence. See also go/gocompiler.vim
func opprec(op ast.Expr) int {
	switch n := op.(type) {
	case *ast.TernaryExpr, *ast.ParenExpr:
		return 1
	case *ast.BinaryExpr:
		switch n.Op {
		case token.OROR:
			return 2
		case token.ANDAND:
			return 3
		case token.EQEQ,
			token.EQEQCI,
			token.EQEQCS,
			token.NEQ,
			token.NEQCI,
			token.NEQCS,
			token.GT,
			token.GTCI,
			token.GTCS,
			token.GTEQ,
			token.GTEQCI,
			token.GTEQCS,
			token.LT,
			token.LTCI,
			token.LTCS,
			token.LTEQ,
			token.LTEQCI,
			token.LTEQCS,
			token.MATCHCS,
			token.NOMATCH,
			token.NOMATCHCI,
			token.NOMATCHCS,
			token.IS,
			token.ISCI,
			token.ISCS,
			token.ISNOT,
			token.ISNOTCI,
			token.ISNOTCS:
			return 4
		case token.PLUS, token.MINUS, token.DOT:
			return 5
		case token.STAR, token.SLASH, token.PERCENT:
			return 6
		default:
			panic(fmt.Errorf("unexpected token of BinaryExpr: %v", n.Op))
		}
	case *ast.UnaryExpr:
		switch n.Op {
		case token.NOT, token.MINUS, token.PLUS:
			return 7
		default:
			panic(fmt.Errorf("unexpected token of UnaryExpr: %v", n.Op))
		}
	case *ast.SubscriptExpr, *ast.SliceExpr, *ast.CallExpr, *ast.DotExpr:
		return 8
	case *ast.BasicLit, *ast.Ident, *ast.List, *ast.Dict, *ast.CurlyName:
		return 9
	case *ast.CurlyNameExpr, *ast.CurlyNameLit, *ast.LambdaExpr:
		panic(fmt.Errorf("precedence is undefined for expr: %T", n))
	default:
		panic(fmt.Errorf("unexpected expr: %T", n))
	}
}
