package types

import "context"

// Expr has an operation and its arguments
type Expr struct {
	fun     Func
	args    []Value
	retType Type
}

// Func returns function of Expr
func (expr *Expr) Func() Func {
	return expr.fun
}

// Args returns arguments of Expr
func (expr *Expr) Args() []Value {
	return expr.args
}

// RetType returns return type of Expr
func (expr *Expr) RetType() Type {
	return expr.retType
}

// NewExpr creates Expr instance
func NewExpr(fun Func, args []Value, retType Type) *Expr {
	return &Expr{fun: fun, args: args, retType: retType}
}

// Eval evaluates given expression expr with given transaction ID trxID.
func (expr *Expr) Eval(ctx context.Context) (val Value, rollback func(), err error) {
	return expr.fun.Execute(ctx, expr.args)
}

// Invert inverts this expression.
// This just calls Func.InvertExpr() with arguments.
func (expr *Expr) Invert() (Value, error) {
	return expr.fun.InvertExpr(expr.args)
}

// Type returns the type.
func (expr *Expr) Type() Type {
	return expr.retType
}
