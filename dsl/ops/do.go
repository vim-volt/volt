package ops

import (
	"context"

	"github.com/vim-volt/volt/dsl/types"
)

func init() {
	opName := doOp("do")
	DoOp = &opName
	opsMap["do"] = DoOp
}

type doOp string

// DoOp is "do" operation
var DoOp *doOp

func (*doOp) String() string {
	return string(*DoOp)
}

func (*doOp) IsMacro() bool {
	return false
}

func (*doOp) Bind(args ...types.Value) (*types.Expr, error) {
	sig := make([]types.Type, 0, len(args))
	for i := 0; i < len(args); i++ {
		sig = append(sig, types.AnyValue)
	}
	if err := signature(sig...).check(args); err != nil {
		return nil, err
	}
	retType := args[len(args)-1].Type()
	return types.NewExpr(DoOp, args, retType), nil
}

func (*doOp) InvertExpr(args []types.Value) (types.Value, error) {
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
