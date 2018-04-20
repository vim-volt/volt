package op

import (
	"github.com/vim-volt/volt/dsl/types"
)

func init() {
	macroMap[string(ArrayOp)] = &ArrayOp
}

type arrayOp string

// ArrayOp is "@" operation
var ArrayOp arrayOp = "@"

// String returns "@"
func (*arrayOp) String() string {
	return string(ArrayOp)
}

// Execute executes "@" operation
func (*arrayOp) Expand(args []types.Value) (types.Value, error) {
	return types.NewArray(args, types.AnyValue), nil
}
