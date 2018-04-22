package util

import (
	"fmt"
	"github.com/pkg/errors"
)

// Guard invokes "rollback functions" if Rollback method received non-nil value
// (e.g. recover(), non-nil error).
type Guard interface {
	// Rollback rolls back if v is non-nil.
	//
	//   defer func() { err = g.Rollback(recover()) }()
	//
	//   // or
	//
	//   if e != nil {
	//     err = g.Rollback(e)
	//     err = g.Rollback(e) // this won't call rollback functions twice!
	//     return
	//   }
	Rollback(v interface{}) error

	// RollbackForcefully calls rollback functions in reversed order
	RollbackForcefully()

	// Add adds given rollback functions
	Add(f func())
}

// FuncGuard returns Guard instance for function
func FuncGuard(name string) Guard {
	return &guard{errMsg: fmt.Sprintf("function \"%s\" has an error", name)}
}

type guard struct {
	errMsg  string
	rbFuncs []func()
}

func (g *guard) Rollback(v interface{}) error {
	var err error
	if e, ok := v.(error); ok {
		err = e
	} else if v != nil {
		err = fmt.Errorf("%s", v)
	}
	if err != nil {
		g.RollbackForcefully()
	}
	return errors.Wrap(err, g.errMsg)
}

func (g *guard) RollbackForcefully() {
	for i := len(g.rbFuncs) - 1; i >= 0; i-- {
		g.rbFuncs[i]()
	}
	g.rbFuncs = nil // do not rollback twice
}

func (g *guard) Add(f func()) {
	g.rbFuncs = append(g.rbFuncs, f)
}
