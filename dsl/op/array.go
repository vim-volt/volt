package op

import (
	"github.com/vim-volt/volt/dsl/types"
)

func init() {
	s := arrayOp("$array")
	ArrayOp = &s
	macroMap[string(*ArrayOp)] = ArrayOp
}

type arrayOp string

// ArrayOp is "$array" operation
var ArrayOp *arrayOp

// String returns "$array"
func (*arrayOp) String() string {
	return string(*ArrayOp)
}

// Execute executes "$array" operation
func (*arrayOp) Expand(args []types.Value) (types.Value, func(), error) {
	return types.NewArray(args, types.AnyValue), noRollback, nil
}
