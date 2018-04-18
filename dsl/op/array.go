package op

import (
	"context"

	"github.com/vim-volt/volt/dsl/types"
)

func init() {
	opsMap[string(ArrayOp)] = &ArrayOp
}

type arrayOp string

// ArrayOp is "@" operation
var ArrayOp arrayOp = "@"

// String returns operator name
func (op *arrayOp) String() string {
	return string(*op)
}

// Bind binds its arguments
func (op *arrayOp) Bind(args ...types.Value) (*types.Expr, error) {
	return &types.Expr{
		Op:   &ArrayOp,
		Args: args,
		Typ:  types.ArrayType,
	}, nil
}

// InvertExpr returns inverted expression
func (op *arrayOp) InvertExpr(args []types.Value) (*types.Expr, error) {
	newargs := make([]types.Value, 0, len(args))
	for i := range args {
		a, err := args[i].Invert()
		if err != nil {
			return nil, err
		}
		newargs = append(newargs, a)
	}
	return ArrayOp.Bind(newargs...)
}

// Execute executes "@" operation
func (op *arrayOp) Execute(ctx context.Context, args []types.Value) (types.Value, func(), error) {
	return &types.Array{Value: args}, func() {}, nil
}
