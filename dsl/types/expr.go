package types

import "context"

// Expr has an operation and its arguments
type Expr struct {
	Func Func
	Args []Value
	Typ  Type
}

// TrxID is a transaction ID, which is a serial number and directory name of
// transaction log file.
// XXX: this should be in transaction package?
type TrxID int64

// Eval evaluates given expression expr with given transaction ID trxID.
func (expr *Expr) Eval(ctx context.Context) (val Value, rollback func(), err error) {
	return expr.Func.Execute(ctx, expr.Args)
}

// Invert inverts this expression.
// This just calls Func.InvertExpr() with arguments.
func (expr *Expr) Invert() (Value, error) {
	return expr.Func.InvertExpr(expr.Args)
}

// Type returns the type.
func (expr *Expr) Type() Type {
	return expr.Typ
}
