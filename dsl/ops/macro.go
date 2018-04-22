package ops

import (
	"context"

	"github.com/vim-volt/volt/dsl/types"
)

type macroBase string

func (m *macroBase) String() string {
	return string(*m)
}

func (*macroBase) IsMacro() bool {
	return true
}

// macroInvertExpr inverts the result of op.Execute() which expands an expression
func (*macroBase) macroInvertExpr(ctx context.Context, val types.Value, _ func(), err error) (types.Value, error) {
	if err != nil {
		return nil, err
	}
	return val.Invert(ctx)
}
