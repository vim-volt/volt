package op

import (
	"github.com/vim-volt/volt/dsl/types"
)

func init() {
	s := invertOp("$invert")
	InvertOp = &s
	macroMap[string(*InvertOp)] = InvertOp
}

type invertOp string

// InvertOp is "$invert" operation
var InvertOp *invertOp

// String returns "$invert"
func (*invertOp) String() string {
	return string(*InvertOp)
}

// Execute executes "$invert" operation
func (*invertOp) Expand(args []types.Value) (types.Value, error) {
	if err := signature(types.AnyValue).check(args); err != nil {
		return nil, err
	}
	return args[0].Invert()
}
