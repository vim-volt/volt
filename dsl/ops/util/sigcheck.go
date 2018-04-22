package util

import (
	"fmt"

	"github.com/vim-volt/volt/dsl/types"
)

// SigChecker checks if the type of args met given types to Signature()
type SigChecker interface {
	Check(args []types.Value) error
}

// Signature returns SigChecker for given types
func Signature(argTypes ...types.Type) SigChecker {
	return &sigChecker{argTypes: argTypes}
}

type sigChecker struct {
	argTypes []types.Type
}

func (sc *sigChecker) Check(args []types.Value) error {
	if len(args) != len(sc.argTypes) {
		return fmt.Errorf("expected %d arity but got %d", len(sc.argTypes), len(args))
	}
	for i := range sc.argTypes {
		if !args[i].Type().InstanceOf(sc.argTypes[i]) {
			return fmt.Errorf("expected %s instance but got %s",
				sc.argTypes[i].String(), args[i].Type().String())
		}
	}
	return nil
}
