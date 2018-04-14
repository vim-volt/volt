package op

import (
	"context"

	"github.com/vim-volt/volt/dsl/types"
)

func init() {
	opsMap["do"] = &DoOp{}
}

// DoOp is "do" operation
type DoOp struct{}

// Describe describes its task(s) as zero or more lines of messages.
func (op *DoOp) Describe(args []types.Value) []string {
	// TODO
	return []string{}
}

// Bind binds its arguments, and check if the types of values are correct.
func (op *DoOp) Bind(args ...types.Value) (*types.Expr, error) {
	sig := make([]types.Type, 0, len(args))
	for i := 0; i < len(args); i++ {
		sig = append(sig, types.ArrayType)
	}
	if err := signature(sig...).check(args); err != nil {
		return nil, err
	}
	retType := args[len(args)-1].Type()
	return types.NewExpr(op, args, retType), nil
}

// InvertExpr returns inverted expression: Call Value.Invert() for each argument,
// and reverse arguments order.
func (op *DoOp) InvertExpr(args []types.Value) (*types.Expr, error) {
	newargs := make([]types.Value, len(args))
	newargs[0] = args[0] // message
	for i := 1; i < len(args); i++ {
		a, err := args[i].Invert()
		if err != nil {
			return nil, err
		}
		newargs[len(args)-i] = a
	}
	return op.Bind(newargs...)
}

// Execute executes "do" operation
func (op *DoOp) Execute(ctx context.Context, args []types.Value) (types.Value, func(), error) {
	// TODO
	return nil, func() {}, nil
}
