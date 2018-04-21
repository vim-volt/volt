package op

import (
	"context"

	"github.com/vim-volt/volt/dsl/types"
)

func init() {
	s := expandMacroOp("$expand-macro")
	ExpandMacroOp = &s
	macroMap[string(*ExpandMacroOp)] = ExpandMacroOp
}

type expandMacroOp string

// ExpandMacroOp is "$expand-macro" operation
var ExpandMacroOp *expandMacroOp

// String returns "$expand-macro"
func (*expandMacroOp) String() string {
	return string(*ArrayOp)
}

// Execute executes "$expand-macro" operation
func (*expandMacroOp) Expand(args []types.Value) (types.Value, func(), error) {
	if err := signature(types.AnyValue).check(args); err != nil {
		return nil, noRollback, err
	}
	return args[0].Eval(context.Background())
}
