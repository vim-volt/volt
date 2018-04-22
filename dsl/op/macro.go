package op

import (
	"github.com/vim-volt/volt/dsl/types"
)

type macroBase string

// String returns "$array"
func (m *macroBase) String() string {
	return string(*m)
}

func (*macroBase) IsMacro() bool {
	return true
}

// Invert the result of op.Execute() which expands an expression
func (*macroBase) macroInvertExpr(val types.Value, _ func(), err error) (types.Value, error) {
	if err != nil {
		return nil, err
	}
	return val.Invert()
}
