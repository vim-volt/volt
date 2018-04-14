package dsl

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/vim-volt/volt/dsl/op"
	"github.com/vim-volt/volt/dsl/types"
)

// ParseOp parses expr JSON.
// This calls Parse() function. And
// 1. Does split to operation and its arguments
// 2. Analyzes semantically its arguments recursively
// 3. Convert the result value to *Expr
func ParseOp(expr []byte) (*types.Expr, error) {
	value, err := Parse(expr)
	if err != nil {
		return nil, err
	}
	array, ok := value.(*types.Array)
	if !ok {
		return nil, errors.New("top-level value is not an array")
	}
	if len(array.Value) == 0 {
		return nil, errors.New("expected operation but got an empty array")
	}
	opName, ok := array.Value[0].(*types.String)
	if !ok {
		return nil, fmt.Errorf("expected operation name but got '%+v'", array.Value[0])
	}
	op, exists := op.Lookup(opName.Value)
	if !exists {
		return nil, fmt.Errorf("no such operation '%s'", opName.Value)
	}
	args := array.Value[1:]
	return op.Bind(args...)
}

// Parse parses expr JSON.
// This only maps encoding/json's types to Value types.
func Parse(expr []byte) (types.Value, error) {
	var value interface{}
	if err := json.Unmarshal(expr, value); err != nil {
		return nil, err
	}
	return convert(value)
}

func convert(value interface{}) (types.Value, error) {
	switch val := value.(type) {
	case nil:
		return types.NullValue, nil
	case bool:
		if val {
			return types.TrueValue, nil
		}
		return types.FalseValue, nil
	case string:
		return &types.String{val}, nil
	case float64:
		return &types.Number{val}, nil
	case []interface{}:
		a := make([]types.Value, 0, len(val))
		for o := range a {
			v, err := convert(o)
			if err != nil {
				return nil, err
			}
			a = append(a, v)
		}
		return &types.Array{a}, nil
	case map[string]interface{}:
		m := make(map[string]types.Value, len(val))
		for k, o := range m {
			v, err := convert(o)
			if err != nil {
				return nil, err
			}
			m[k] = v
		}
		return &types.Object{m}, nil
	default:
		return nil, fmt.Errorf("unknown value was given '%+v'", val)
	}
}
