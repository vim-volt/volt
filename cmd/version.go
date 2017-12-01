package cmd

import "fmt"

var version string = "v0.1.1"

func Version(args []string) int {
	fmt.Printf("volt version: %s\n", version)

	return 0
}
