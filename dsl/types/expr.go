package types

import "context"

// Expr has an operation and its arguments
type Expr struct {
	op   Op
	args []Value
	typ  Type
}

// NewExpr is the constructor of Expr
func NewExpr(op Op, args []Value, typ Type) *Expr {
	return &Expr{op: op, args: args, typ: typ}
}

// TrxID is a transaction ID, which is a serial number and directory name of
// transaction log file.
// XXX: this should be in transaction package?
type TrxID int64

// Eval evaluates given expression expr with given transaction ID trxID.
func (expr *Expr) Eval(ctx context.Context) (val Value, rollback func(), err error) {
	return expr.op.Execute(ctx, expr.args)
}

// Invert inverts this expression.
// This just calls Op.InvertExpr() with arguments.
func (expr *Expr) Invert() (Value, error) {
	return expr.op.InvertExpr(expr.args)
}

// Describe describes its task(s) as zero or more lines of messages.
func (expr *Expr) Describe() []string {
	return expr.op.Describe(expr.args)
}

// Type returns the type.
func (expr *Expr) Type() Type {
	return expr.typ
}
