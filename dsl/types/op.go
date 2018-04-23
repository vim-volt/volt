package types

import "context"

// Op is an operator of JSON DSL
type Op interface {
	// String returns function name
	String() string

	// InvertExpr returns inverted expression
	InvertExpr(ctx context.Context, args []Value) (Value, error)

	// Bind binds its arguments, and check if the types of values are correct
	Bind(args ...Value) (Expr, error)

	// EvalExpr evaluates expression (this operator + given arguments).
	// If this operator is a function, it executes the operation and returns its
	// result and error.
	// If this operator is a macro, this expands expression.
	EvalExpr(ctx context.Context, args []Value) (ret Value, rollback func(context.Context), err error)

	// IsMacro returns true if this operator is a macro
	IsMacro() bool
}
