package util

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// Guard invokes "rollback functions" if Rollback method received non-nil value
// (e.g. recover(), non-nil error).
type Guard interface {
	// Error sets v as an error if v is non-nil.
	// This returns the error.
	//
	//   defer func() {
	//     result = g.Error(recover())
	//   }()
	//
	//   // or
	//
	//   if err != nil {
	//     return g.Error(err)
	//   }
	//
	Error(v interface{}) error

	// Rollback calls rollback functions in reversed order
	Rollback(ctx context.Context)

	// Add adds given rollback functions, but skips if f == nil
	Add(f func(context.Context))
}

// FuncGuard returns Guard instance for function
func FuncGuard(name string) Guard {
	return &guard{errMsg: fmt.Sprintf("function \"%s\" has an error", name)}
}

type guard struct {
	errMsg  string
	err     error
	rbFuncs []func(context.Context)
}

func (g *guard) Error(v interface{}) error {
	if err, ok := v.(error); ok {
		g.err = errors.Wrap(err, g.errMsg)
	} else if v != nil {
		g.err = errors.Wrap(fmt.Errorf("%s", v), g.errMsg)
	}
	return g.err
}

func (g *guard) Rollback(ctx context.Context) {
	for i := len(g.rbFuncs) - 1; i >= 0; i-- {
		g.rbFuncs[i](ctx)
	}
	g.rbFuncs = nil // do not rollback twice
}

func (g *guard) Add(f func(context.Context)) {
	if f != nil {
		g.rbFuncs = append(g.rbFuncs, f)
	}
}
