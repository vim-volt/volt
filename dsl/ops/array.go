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

func (op *arrayOp) InvertExpr(args []types.Value) (types.Value, error) {
	return op.macroInvertExpr(op.Execute(context.Background(), args))
}

func (*arrayOp) Bind(args ...types.Value) (*types.Expr, error) {
	expr := types.NewExpr(ArrayOp, args, types.NewArrayType(types.AnyValue))
	return expr, nil
}

func (*arrayOp) Execute(ctx context.Context, args []types.Value) (types.Value, func(), error) {
	return types.NewArray(args, types.AnyValue), NoRollback, nil
}
