package op

import (
	"context"

	"github.com/vim-volt/volt/dsl/types"
)

func init() {
	s := evalOp("$eval")
	EvalOp = &s
	macroMap[string(*EvalOp)] = EvalOp
}

type evalOp string

// EvalOp is "$eval" operation
var EvalOp *evalOp

// String returns "$eval"
func (*evalOp) String() string {
	return string(*EvalOp)
}

// Execute executes "$eval" operation
func (*evalOp) Expand(args []types.Value) (types.Value, func(), error) {
	if err := signature(types.AnyValue).check(args); err != nil {
		return nil, NoRollback, err
	}
	return args[0].Eval(context.Background())
}
