package types

import "context"

// Value is JSON value
type Value interface {
	// Invert returns inverted value/operation.
	// All type values are invertible.
	// Literals like string,number,... return itself as-is.
	// If argument type or arity is different, this returns non-nil error.
	Invert() (Value, error)

	// Eval returns a evaluated value.
	// Literals like string,number,... return itself as-is.
	Eval(ctx context.Context) (val Value, rollback func(), err error)

	// Type returns the type of this value.
	Type() Type
}
