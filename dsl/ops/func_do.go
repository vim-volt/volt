package ops

import (
	"context"

	"github.com/vim-volt/volt/dsl/ops/util"
	"github.com/vim-volt/volt/dsl/types"
)

func init() {
	opsMap[DoOp.String()] = DoOp
}

type doOp struct {
	funcBase
}

// DoOp is "do" operation
var DoOp = &doOp{funcBase("do")}

func (*doOp) Bind(args ...types.Value) (types.Expr, error) {
	sig := make([]types.Type, 0, len(args))
	for i := 0; i < len(args); i++ {
		sig = append(sig, types.AnyValue)
	}
	if err := util.Signature(sig...).Check(args); err != nil {
		return nil, err
	}
	retType := args[len(args)-1].Type()
	return types.NewExpr(DoOp, args, retType), nil
}

func (*doOp) InvertExpr(ctx context.Context, args []types.Value) (types.Value, error) {
	newargs := make([]types.Value, len(args))
	for i := range args {
		a, err := args[i].Invert(ctx)
		if err != nil {
			return nil, err
		}
		newargs[len(args)-i] = a
	}
	return DoOp.Bind(newargs...)
}

func (*doOp) EvalExpr(ctx context.Context, args []types.Value) (_ types.Value, _ func(), result error) {
	g := util.FuncGuard(DoOp.String())
	defer func() {
		result = g.Error(recover())
	}()

	var lastVal types.Value
	for i := range args {
		v, rbFunc, err := args[i].Eval(ctx)
		g.Add(rbFunc)
		if err != nil {
			result = g.Error(err)
			return
		}
		lastVal = v
	}
	return lastVal, g.Rollback, nil
}
