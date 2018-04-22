package dsl

import (
	"encoding/json"
	"fmt"

	"github.com/vim-volt/volt/dsl/ops"
	"github.com/vim-volt/volt/dsl/types"
)

// Deparse deparses types.Expr.
// ["@", 1, 2, 3] becomes [1, 2, 3]
func Deparse(expr types.Expr) (interface{}, error) {
	value, err := deparse(expr)
	if err != nil {
		return nil, err
	}
	return json.Marshal(value)
}

func deparse(value types.Value) (interface{}, error) {
	if _, ok := value.Type().(*types.NullType); ok {
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
		m := make(map[string]interface{}, len(val.Value()))
		for k, o := range val.Value() {
			v, err := deparse(o)
			if err != nil {
				return nil, err
			}
			m[k] = v
		}
		return m, nil
	case types.Expr:
		a := make([]interface{}, 0, len(val.Args())+1)
		// Do not include "@" in array literal
		if val.Op().String() != ops.ArrayOp.String() {
			a = append(a, types.NewString(val.Op().String()))
		}
		for i := range a {
			v, err := deparse(val.Args()[i])
			if err != nil {
				return nil, err
			}
			a = append(a, v)
		}
		return a, nil
	default:
		return nil, fmt.Errorf("unknown value was given '%+v'", val)
	}
}
