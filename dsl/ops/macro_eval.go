package ops

import (
	"context"

	"github.com/vim-volt/volt/dsl/ops/util"
	"github.com/vim-volt/volt/dsl/types"
)

func init() {
	opsMap[EvalOp.String()] = EvalOp
}

type evalOp struct {
	macroBase
}

// EvalOp is "$eval" operator
var EvalOp = &evalOp{macroBase("$eval")}

func (op *evalOp) InvertExpr(ctx context.Context, args []types.Value) (types.Value, error) {
	val, rollback, err := op.EvalExpr(ctx, args)
	return op.macroInvertExpr(ctx, val, rollback, err)
}

func (*evalOp) Bind(args ...types.Value) (types.Expr, error) {
	expr := types.NewExpr(ArrayOp, args, types.NewArrayType(types.AnyValue))
	return expr, nil
}

func (*evalOp) EvalExpr(ctx context.Context, args []types.Value) (types.Value, func(), error) {
	if err := util.Signature(types.AnyValue).Check(args); err != nil {
		return nil, nil, err
	}
	return args[0].Eval(ctx)
}
