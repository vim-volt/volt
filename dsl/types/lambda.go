package types

import (
	"context"

	"github.com/pkg/errors"
)

// Lambda can be applicable, and it has an expression to execute.
type Lambda interface {
	Value

	// Call calls this lambda with given args
	Call(ctx context.Context, args ...Value) (Value, func(context.Context), error)
}

type argT Array

// ArgsDef is passed to builder function, the argument of NewLambda()
type ArgsDef struct {
	args []argT
}

// Define returns placeholder expression of given argument
func (def *ArgsDef) Define(n int, name String, typ Type) (Value, error) {
	if n <= 0 {
		return nil, errors.New("the number of argument must be positive")
	}
	for n > len(def.args) {
		def.args = append(def.args, nil)
	}
	if def.args[n-1] != nil {
		return nil, errors.Errorf("the %dth argument is already taken", n)
	}
	argExpr := []Value{NewString("arg"), name}
	def.args[n-1] = NewArray(argExpr, AnyValue)
	return def.args[n-1], nil
}

// Inject replaces expr of ["arg", expr] with given values
func (def *ArgsDef) Inject(args []Value) error {
	if len(args) != len(def.args) {
		return errors.Errorf("expected %d arity but got %d", len(def.args), len(args))
	}
	for i := range args {
		if def.args[i] == nil {
			return errors.Errorf("%dth arg is not taken", i+1)
		}
		def.args[i].Value()[1] = args[i]
	}
	return nil
}

// NewLambda creates lambda value.
// Signature must have 1 type at least for a return type.
func NewLambda(builder func(*ArgsDef) (Expr, []Type, error)) (Lambda, error) {
	def := &ArgsDef{args: make([]argT, 0)}
	expr, sig, err := builder(def)
	if err != nil {
		return nil, errors.Wrap(err, "builder function returned an error")
	}
	return &lambdaT{
		def:  def,
		expr: expr,
		typ:  NewLambdaType(sig[0], sig[1:]...),
	}, nil
}

type lambdaT struct {
	def  *ArgsDef
	expr Expr
	typ  Type
}

func (v *lambdaT) Call(ctx context.Context, args ...Value) (Value, func(context.Context), error) {
	if err := v.def.Inject(args); err != nil {
		return nil, nil, err
	}
	return v.expr.Eval(ctx)
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
