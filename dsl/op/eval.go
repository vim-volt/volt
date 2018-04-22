package op

import (
	"context"

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

func (op *evalOp) InvertExpr(args []types.Value) (types.Value, error) {
	return op.macroInvertExpr(op.Execute(context.Background(), args))
}

func (*evalOp) Bind(args ...types.Value) (*types.Expr, error) {
	expr := types.NewExpr(ArrayOp, args, types.NewArrayType(types.AnyValue))
	return expr, nil
}

func (*evalOp) Execute(ctx context.Context, args []types.Value) (types.Value, func(), error) {
	if err := signature(types.AnyValue).check(args); err != nil {
		return nil, NoRollback, err
	}
	return args[0].Eval(context.Background())
}
