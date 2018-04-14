package dsl

import (
	"context"
	"errors"

	"github.com/vim-volt/volt/dsl/types"
)

// CtxKeyType is the type of the key of context specified for Execute()
type CtxKeyType uint

const (
	// CtxTrxIDKey is the key to get transaction ID
	CtxTrxIDKey CtxKeyType = iota
	// CtxLockJSONKey is the key to get *lockjson.LockJSON value
	CtxLockJSONKey
	// CtxConfigKey is the key to get *config.Config value
	CtxConfigKey
)

// Execute executes given expr with given ctx.
func Execute(ctx context.Context, expr *types.Expr) (val types.Value, rollback func(), err error) {
	ctx = context.WithValue(ctx, CtxTrxIDKey, genNewTrxID())
	if ctx.Value(CtxLockJSONKey) == nil {
		return nil, func() {}, errors.New("no lock.json key in context")
	}
	if ctx.Value(CtxConfigKey) == nil {
		return nil, func() {}, errors.New("no config.toml key in context")
	}
	return expr.Eval(ctx)
}

func genNewTrxID() types.TrxID {
	// TODO: Get unallocated transaction ID looking $VOLTPATH/trx/ directory
	return types.TrxID(0)
}
