package main

import (
	"os"

	"github.com/vim-volt/volt/cmd"
	"github.com/vim-volt/volt/logger"
)

func main() {
	os.Exit(Main())
}

func Main() int {
	if os.Getenv("VOLT_DEBUG") != "" {
		logger.SetLevel(logger.DebugLevel)
	}
	if len(os.Args) <= 1 {
		os.Args = append(os.Args, "help")
	}
	switch os.Args[1] {
	case "get":
		return cmd.Get(os.Args[2:])
	case "rm":
		return cmd.Rm(os.Args[2:])
	case "enable":
		return cmd.Enable(os.Args[2:])
	case "disable":
		return cmd.Disable(os.Args[2:])
	case "list":
		return cmd.List(os.Args[2:])
	case "profile":
		return cmd.Profile(os.Args[2:])
	case "build":
		return cmd.Build(os.Args[2:])
	case "migrate":
		return cmd.Migrate(os.Args[2:])
	case "self-upgrade":
		return cmd.SelfUpgrade(os.Args[2:])
	case "version":
		return cmd.Version(os.Args[2:])
	case "help":
		return cmd.Help(os.Args[2:])
	default:
		logger.Error("Unknown command '" + os.Args[1] + "'")
		return 3
	}
}
