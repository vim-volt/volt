package ops

import (
	"context"

	"github.com/pkg/errors"
	"github.com/vim-volt/volt/dsl/dslctx"
	"github.com/vim-volt/volt/dsl/ops/util"
	"github.com/vim-volt/volt/dsl/types"
	"github.com/vim-volt/volt/lockjson"
)

func init() {
	opsMap[LockJSONWriteOp.String()] = LockJSONWriteOp
}

type lockJSONWriteOp struct {
	funcBase
}

// LockJSONWriteOp is "lockjson/write" operator
var LockJSONWriteOp = &lockJSONWriteOp{funcBase("lockjson/write")}

func (*lockJSONWriteOp) Bind(args ...types.Value) (types.Expr, error) {
	if err := util.Signature().Check(args); err != nil {
		return nil, err
	}
	retType := types.VoidType
	return types.NewExpr(LockJSONWriteOp, args, retType), nil
}

func (*lockJSONWriteOp) InvertExpr(_ context.Context, args []types.Value) (types.Value, error) {
	return LockJSONWriteOp.Bind(args...)
}

func (*lockJSONWriteOp) EvalExpr(ctx context.Context, args []types.Value) (_ types.Value, rollback func(), result error) {
	rollback = NoRollback

	lockJSON := ctx.Value(dslctx.LockJSONKey).(*lockjson.LockJSON)
	result = lockJSON.Write()
	if result != nil {
		result = errors.Wrap(result, "could not write to lock.json")
	}

	return
}
