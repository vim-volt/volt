package op

import (
	"context"

	"github.com/vim-volt/volt/dsl/types"
)

func init() {
	funcMap[string(DoOp)] = &DoOp
}

type doOp string

// DoOp is "do" operation
var DoOp doOp = "do"

// String returns operator name
func (*doOp) String() string {
	return string(DoOp)
}

// Bind binds its arguments, and check if the types of values are correct.
func (*doOp) Bind(args ...types.Value) (*types.Expr, error) {
	sig := make([]types.Type, 0, len(args))
	for i := 0; i < len(args); i++ {
		sig = append(sig, types.AnyValue)
	}
	if err := signature(sig...).check(args); err != nil {
		return nil, err
	}
	retType := args[len(args)-1].Type()
	return types.NewExpr(&DoOp, args, retType), nil
}

// InvertExpr returns inverted expression: Call Value.Invert() for each argument,
// and reverse arguments order.
func (*doOp) InvertExpr(args []types.Value) (*types.Expr, error) {
	newargs := make([]types.Value, len(args))
	for i := range args {
		a, err := args[i].Invert()
		if err != nil {
			return nil, err
		}
		newargs[len(args)-i] = a
	}
	return DoOp.Bind(newargs...)
}

// Execute executes "do" operation
func (*doOp) Execute(ctx context.Context, args []types.Value) (val types.Value, rollback func(), err error) {
	g := funcGuard(DoOp.String())
	defer func() { err = g.rollback(recover()) }()
	rollback = g.rollbackForcefully

	for i := range args {
		v, rbFunc, e := args[i].Eval(ctx)
		g.add(rbFunc)
		if e != nil {
			err = g.rollback(e)
			return
		}
		val = v
	}
	return
}
