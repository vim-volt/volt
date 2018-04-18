package types

import "context"

// Op is an operation of JSON DSL
type Op interface {
	// Bind binds its arguments, and check if the types of values are correct.
	Bind(args ...Value) (*Expr, error)

	// InvertExpr returns inverted expression
	InvertExpr(args []Value) (*Expr, error)

	// Execute executes this operation and returns its result value and error
	Execute(ctx context.Context, args []Value) (ret Value, rollback func(), err error)

	// Returns operator name
	String() string
}
