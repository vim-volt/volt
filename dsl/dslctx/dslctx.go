package dslctx

import (
	"context"
	"errors"

	"github.com/vim-volt/volt/config"
	"github.com/vim-volt/volt/lockjson"
)

// KeyType is the type of the key of context specified for Execute()
type KeyType uint

const (
	// TrxIDKey is the key to get transaction ID
	TrxIDKey KeyType = iota
	// LockJSONKey is the key to get *lockjson.LockJSON value
	LockJSONKey
	// ConfigKey is the key to get *config.Config value
	ConfigKey
)

// WithDSLValues adds given values
func WithDSLValues(ctx context.Context, lockJSON *lockjson.LockJSON, cfg *config.Config) context.Context {
	ctx = context.WithValue(ctx, LockJSONKey, lockJSON)
	ctx = context.WithValue(ctx, ConfigKey, cfg)
	return ctx
}

// Validate validates if required keys exist in ctx
func Validate(ctx context.Context) error {
	for _, required := range []struct {
		key      KeyType
		validate func(interface{}) error
	}{
		{LockJSONKey, validateLockJSON},
		{ConfigKey, validateConfig},
	} {
		if err := required.validate(ctx.Value(required.key)); err != nil {
			return err
		}
	}
	return nil
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
