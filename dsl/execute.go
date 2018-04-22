package dsl

import (
	"context"

	"github.com/pkg/errors"
	"github.com/vim-volt/volt/dsl/dslctx"
	"github.com/vim-volt/volt/dsl/types"
	"github.com/vim-volt/volt/transaction"
)

// Execute executes given expr with given ctx.
func Execute(ctx context.Context, expr types.Expr) (_ types.Value, result error) {
	if err := dslctx.Validate(ctx); err != nil {
		return nil, err
	}

	// Begin transaction
	trx, err := transaction.Start()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := trx.Done(); err != nil {
			result = err
		}
	}()

	val, rollback, err := expr.Eval(ctx)
	if err != nil {
		rollback()
		return nil, errors.Wrap(err, "expression returned an error")
	}
	return val, nil
}
