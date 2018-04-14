// +build go1.9

package main

import (
	"os"

	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/subcmd"
)

func main() {
	err := subcmd.Run(os.Args, subcmd.DefaultRunner)
	if err != nil {
		logger.Error(err.Msg)
		os.Exit(err.Code)
	}
}
