// +build go1.9

package main

import (
	"os"

	"github.com/rjkat/volt/logger"
	"github.com/rjkat/volt/subcmd"
)

func main() {
	err := subcmd.Run(os.Args, subcmd.DefaultRunner)
	if err != nil {
		logger.Error(err.Msg)
		os.Exit(err.Code)
	}
}
