// +build go1.9

package main

import (
	"os"

	"github.com/vim-volt/volt/gateway"
	"github.com/vim-volt/volt/logger"
)

func main() {
	err := gateway.Run(os.Args, gateway.DefaultRunner)
	if err != nil {
		logger.Error(err.Msg)
		os.Exit(err.Code)
	}
}
