package op

import "github.com/vim-volt/volt/dsl/types"

// funcMap holds all operation structs.
// All operations in dsl/op/*.go sets its struct to this in init()
var funcMap map[string]types.Func

// LookupFunc looks up function name
func LookupFunc(name string) (types.Func, bool) {
	op, exists := funcMap[name]
	return op, exists
}

// macroMap holds all operation structs.
// All operations in dsl/op/*.go sets its struct to this in init()
var macroMap map[string]types.Macro

// LookupMacro looks up macro name
func LookupMacro(name string) (types.Macro, bool) {
	op, exists := macroMap[name]
	return op, exists
}
