package op

import (
	"fmt"
	"github.com/pkg/errors"
)

// guard invokes "rollback functions" if rollback method received non-nil value
// (e.g. recover(), non-nil error).
type guard struct {
	errMsg  string
	rbFuncs []func()
}

// rollback rolls back if v is non-nil.
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
func (g *guard) rollback(v interface{}) error {
	var err error
	if e, ok := v.(error); ok {
		err = e
	} else if v != nil {
		err = fmt.Errorf("%s", v)
	}
	if err != nil {
		g.rollbackForcefully()
	}
	return errors.Wrap(err, g.errMsg)
}

// rollbackForcefully calls rollback functions in reversed order
func (g *guard) rollbackForcefully() {
	for i := len(g.rbFuncs) - 1; i >= 0; i-- {
		g.rbFuncs[i]()
	}
	g.rbFuncs = nil // do not rollback twice
}

// add adds given rollback functions
func (g *guard) add(f func()) {
	g.rbFuncs = append(g.rbFuncs, f)
}

func funcGuard(name string) *guard {
	return &guard{errMsg: fmt.Sprintf("function \"%s\" has an error", name)}
}
