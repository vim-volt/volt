package op

import "github.com/vim-volt/volt/dsl/types"

func signature(sig ...types.Type) *sigChecker {
	return &sigChecker{sig}
}

type sigChecker struct {
	sig []types.Type
}

func (sc *sigChecker) check(args []types.Value) error {
	// TODO
	return nil
}
