package cmd

import (
	"flag"

	"github.com/vim-volt/volt/logger"
)

var cmdMap = make(map[string]Cmd)

type Cmd interface {
	Run(args []string) int
	FlagSet() *flag.FlagSet
}

func Run(subCmd string, args []string) int {
	if self, exists := cmdMap[subCmd]; exists {
		return self.Run(args)
	}
	logger.Error("Unknown command '" + subCmd + "'")
	return 3
}
