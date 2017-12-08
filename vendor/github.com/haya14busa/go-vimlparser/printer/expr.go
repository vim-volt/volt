package printer

import (
	"fmt"

	"github.com/haya14busa/go-vimlparser/ast"
	"github.com/haya14busa/go-vimlparser/token"
)

func (p *printer) expr(expr ast.Expr) {
	p.expr1(expr, 0)
}

func (p *printer) expr1(expr ast.Expr, prec1 int) {
	switch x := expr.(type) {
	// case *ast.TernaryExpr:
	case *ast.BinaryExpr:
		p.binaryExpr(x, prec1)
	case *ast.UnaryExpr:
		p.unaryExpr(x, prec1)
	case *ast.SubscriptExpr:
		p.expr1(x.Left, opprec(x))
		p.token(token.SQOPEN)
		p.expr(x.Right)
		p.token(token.SQCLOSE)
	// case *ast.SliceExpr:
	// case *ast.CallExpr:
	// case *ast.DotExpr:
	// case *ast.List:
	// case *ast.Dict:
	// case *ast.CurlyName:
	// case *ast.CurlyNameLit:
	// case *ast.CurlyNameExpr:
	case *ast.BasicLit:
		p.writeString(x.Value)
	case *ast.Ident:
		p.writeString(x.Name)
	// case *ast.LambdaExpr:
	case *ast.ParenExpr:
		if _, hasParens := x.X.(*ast.ParenExpr); hasParens {
			p.expr(x.X)
		} else {
			p.token(token.POPEN)
			p.expr(x.X)
			p.token(token.PCLOSE)
		}
	default:
		panic(fmt.Errorf("unsupported expr type %T", x))
	}
}

func (p *printer) binaryExpr(x *ast.BinaryExpr, prec1 int) {
	prec := opprec(x)
	if prec < prec1 {
		// parenthesis needed
		// Note: The parser inserts an ast.ParenExpr node; thus this case
		//       can only occur if the AST is created in a different way.
		p.token(token.POPEN)
		p.expr(x)
		p.token(token.PCLOSE)
		return
	}
	// TODO(haya14busa): handle line break.
	p.expr1(x.Left, prec)
	p.printWhite(blank)
	p.token(x.Op)
	p.printWhite(blank)
	p.expr1(x.Right, prec+1)
}

func (p *printer) unaryExpr(x *ast.UnaryExpr, prec1 int) {
	prec := opprec(x)
	if prec < prec1 {
		// parenthesis needed
		// Note: this case can only occcur if the AST is created manually and
		// the code should invalid code.
		p.token(token.POPEN)
		p.expr(x)
		p.token(token.PCLOSE)
		return
	}
	// no parenthesis needed
	p.token(x.Op)
	p.expr1(x.X, prec)
}
