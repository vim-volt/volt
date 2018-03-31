// +build go1.9

package main

import (
	"os"

	"github.com/vim-volt/volt/cmd"
	"github.com/vim-volt/volt/logger"
)

func main() {
	os.Exit(doMain())
}

func doMain() int {
	if os.Getenv("VOLT_DEBUG") != "" {
		logger.SetLevel(logger.DebugLevel)
	}
	if len(os.Args) <= 1 {
		os.Args = append(os.Args, "help")
	}
	return cmd.Run(os.Args[1], os.Args[2:])
}
