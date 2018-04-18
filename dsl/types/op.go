package types

import "context"

// Func is an operation of JSON DSL
type Func interface {
	// Returns function name
	String() string

	// InvertExpr returns inverted expression
	InvertExpr(args []Value) (*Expr, error)

	// Bind binds its arguments, and check if the types of values are correct.
	Bind(args ...Value) (*Expr, error)

	// Execute executes this operation and returns its result value and error
	Execute(ctx context.Context, args []Value) (ret Value, rollback func(), err error)
}

// Macro is an operation of JSON DSL
type Macro interface {
	// Returns macro name
	String() string

	// Execute executes this operation and returns its result value and error
	Expand(args []Value) Value
}
