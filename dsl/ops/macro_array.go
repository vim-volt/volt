package ops

import (
	"context"

	"github.com/vim-volt/volt/dsl/types"
)

func init() {
	opsMap[ArrayOp.String()] = ArrayOp
}

type arrayOp struct {
	macroBase
}

// ArrayOp is "$array" operator
var ArrayOp = &arrayOp{macroBase("$array")}

func (op *arrayOp) InvertExpr(ctx context.Context, args []types.Value) (types.Value, error) {
	val, rollback, err := op.EvalExpr(ctx, args)
	return op.macroInvertExpr(ctx, val, rollback, err)
}

func (*arrayOp) Bind(args ...types.Value) (types.Expr, error) {
	expr := types.NewExpr(ArrayOp, args, types.NewArrayType(types.AnyValue))
	return expr, nil
}

func (*arrayOp) EvalExpr(ctx context.Context, args []types.Value) (types.Value, func(), error) {
	return types.NewArray(args, types.AnyValue), NoRollback, nil
}
