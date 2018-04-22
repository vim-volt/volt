package types

import "context"

// Expr has an operation and its arguments
type Expr struct {
	op      Op
	args    []Value
	retType Type
}

// Op returns operator of Expr
func (expr *Expr) Op() Op {
	return expr.op
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
func NewExpr(op Op, args []Value, retType Type) *Expr {
	return &Expr{op: op, args: args, retType: retType}
}

// Eval evaluates given expression expr with given transaction ID trxID.
func (expr *Expr) Eval(ctx context.Context) (val Value, rollback func(), err error) {
	return expr.op.EvalExpr(ctx, expr.args)
}

// Invert inverts this expression.
// This just calls Op().InvertExpr() with saved arguments.
func (expr *Expr) Invert() (Value, error) {
	return expr.op.InvertExpr(expr.args)
}

// Type returns the type.
func (expr *Expr) Type() Type {
	return expr.retType
}
