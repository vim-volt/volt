package dsl

import (
	"context"

	"github.com/pkg/errors"
	"github.com/vim-volt/volt/dsl/op"
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
	for _, st := range []struct {
		key  CtxKeyType
		name string
	}{
		{CtxLockJSONKey, "lock.json"},
		{CtxConfigKey, "config.toml"},
		{CtxTrxIDKey, "transaction ID"},
	} {
		if ctx.Value(st.key) == nil {
			return nil, op.NoRollback, errors.Errorf("no %s key in context", st.name)
		}
	}
	return expr.Eval(ctx)
}
