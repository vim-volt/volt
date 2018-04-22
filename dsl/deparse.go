package dsl

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/vim-volt/volt/dsl/ops"
	"github.com/vim-volt/volt/dsl/types"
)

// Deparse deparses types.Expr.
// ["@", 1, 2, 3] becomes [1, 2, 3]
func Deparse(expr types.Expr) ([]byte, error) {
	value, err := deparse(expr)
	if err != nil {
		return nil, err
	}
	return json.Marshal(value)
}

func deparse(value types.Value) (interface{}, error) {
	if value.Type() == types.NullType {
		return nil, nil
	}
	switch val := value.(type) {
	case types.Bool:
		return val.Value(), nil
	case types.String:
		return val.Value(), nil
	case types.Number:
		return val.Value(), nil
	case types.Object:
		result := make(map[string]interface{}, len(val.Value()))
		for k, o := range val.Value() {
			v, err := deparse(o)
			if err != nil {
				return nil, err
			}
			result[k] = v
		}
		return result, nil
	case types.Expr:
		args := val.Args()
		result := make([]interface{}, 0, len(args)+1)
		// Do not include "@" in array literal
		if val.Op().String() != ops.ArrayOp.String() {
			result = append(result, val.Op().String())
		}
		for i := range args {
			v, err := deparse(args[i])
			if err != nil {
				return nil, err
			}
			result = append(result, v)
		}
		return result, nil
	default:
		return nil, errors.Errorf("unknown value was given '%+v'", val)
	}
}
