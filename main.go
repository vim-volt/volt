// +build go1.9

package main

import (
	"os"

	"github.com/vim-volt/volt/cmd"
	"github.com/vim-volt/volt/logger"
)

func main() {
	err := cmd.Run(os.Args, cmd.DefaultRunner)
	if err != nil {
		logger.Error(err.Msg)
		os.Exit(err.Code)
	}
}
