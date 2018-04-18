package dsl

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/vim-volt/volt/dsl/op"
	"github.com/vim-volt/volt/dsl/types"
)

// Parse parses expr JSON. And if an array literal value is found:
// 1. Split to operation and its arguments
// 2. Do semantic analysis recursively for its arguments
// 3. Convert to *Expr
func Parse(expr []byte) (types.Value, error) {
	var value interface{}
	if err := json.Unmarshal(expr, value); err != nil {
		return nil, err
	}
	return parseExpr(value)
}

func parseExpr(value interface{}) (*types.Expr, error) {
	if array, ok := value.([]interface{}); ok {
		return parseArray(array)
	}
	return nil, errors.New("top-level must be an array")
}

func parseArray(array []interface{}) (*types.Expr, error) {
	if len(array) == 0 {
		return nil, errors.New("expected operation but got an empty array")
	}
	opName, ok := array[0].(string)
	if !ok {
		return nil, fmt.Errorf("expected operator (string) but got '%+v'", array[0])
	}
	a := make([]types.Value, 0, len(array)-1)
	for i := 1; i < len(array); i++ {
		v, err := parse(array[i])
		if err != nil {
			return nil, err
		}
		a = append(a, v)
	}
	op, exists := op.Lookup(opName)
	if !exists {
		return nil, fmt.Errorf("no such operation '%s'", opName)
	}
	return op.Bind(a...)
}

func parse(value interface{}) (types.Value, error) {
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
	case map[string]interface{}:
		m := make(map[string]types.Value, len(val))
		for k, o := range m {
			v, err := parse(o)
			if err != nil {
				return nil, err
			}
			m[k] = v
		}
		return &types.Object{m}, nil
	case []interface{}:
		return parseArray(val)
	default:
		return nil, fmt.Errorf("unknown value was given '%+v'", val)
	}
}
