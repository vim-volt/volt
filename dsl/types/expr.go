package types

import "context"

// Expr has an operation and its arguments
type Expr interface {
	// Op returns operator of Expr
	Op() Op

	// Args returns arguments of Expr
	Args() []Value

	// RetType returns return type of Expr
	RetType() Type

	// Eval evaluates given expression expr with given transaction ID trxID.
	Eval(ctx context.Context) (val Value, rollback func(), err error)

	// Invert inverts this expression.
	// This just calls Op().InvertExpr() with saved arguments.
	Invert() (Value, error)

	// Type returns the type.
	Type() Type
}

// NewExpr creates Expr instance
func NewExpr(op Op, args []Value, retType Type) Expr {
	return &expr{op: op, args: args, retType: retType}
}

type expr struct {
	op      Op
	args    []Value
	retType Type
}

func (expr *expr) Op() Op {
	return expr.op
}

func (expr *expr) Args() []Value {
	return expr.args
}

func (expr *expr) RetType() Type {
	return expr.retType
}

func (expr *expr) Eval(ctx context.Context) (val Value, rollback func(), err error) {
	return expr.op.EvalExpr(ctx, expr.args)
}

func (expr *expr) Invert() (Value, error) {
	return expr.op.InvertExpr(expr.args)
}

func (expr *expr) Type() Type {
	return expr.retType
}
