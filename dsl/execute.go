package dsl

import (
	"context"

	"github.com/pkg/errors"
	"github.com/vim-volt/volt/config"
	"github.com/vim-volt/volt/dsl/ops"
	"github.com/vim-volt/volt/dsl/types"
	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/transaction"
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
	for _, required := range []struct {
		key      CtxKeyType
		validate func(interface{}) error
	}{
		{CtxLockJSONKey, validateLockJSON},
		{CtxConfigKey, validateConfig},
		{CtxTrxIDKey, validateTrxID},
	} {
		if err := required.validate(ctx.Value(required.key)); err != nil {
			return nil, ops.NoRollback, err
		}
	}
	return expr.Eval(ctx)
}

func validateLockJSON(v interface{}) error {
	if v == nil {
		return errors.New("no lock.json key in context")
	}
	if _, ok := v.(*lockjson.LockJSON); !ok {
		return errors.New("invalid lock.json data in context")
	}
	return nil
}

func validateConfig(v interface{}) error {
	if v == nil {
		return errors.New("no config.toml key in context")
	}
	if _, ok := v.(*config.Config); !ok {
		return errors.New("invalid config.toml data in context")
	}
	return nil
}

func validateTrxID(v interface{}) error {
	if v == nil {
		return errors.New("no transaction ID key in context")
	}
	if _, ok := v.(transaction.TrxID); !ok {
		return errors.New("invalid transaction ID data in context")
	}
	return nil
}
