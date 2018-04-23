package types

import "context"

// Lambda can be applicable, and it has an expression to execute.
type Lambda Value

// NewLambda creates lambda value.
// Signature must have 1 type at least for a return type.
func NewLambda(t Type, rest ...Type) Lambda {
	return &lambdaT{typ: NewLambdaType(t, rest...)}
}

type lambdaT struct {
	typ Type
}

func (v *lambdaT) Invert(context.Context) (Value, error) {
	return v, nil
}

func (v *lambdaT) Eval(context.Context) (Value, func(context.Context), error) {
	return v, nil, nil
}

func (v *lambdaT) Type() Type {
	return v.typ
}
