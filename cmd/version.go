package cmd

import "fmt"

var version string = "v0.0.3"
var revision string = "Devel"

func Version(args []string) int {
	fmt.Printf("volt version: %s (rev %s)\n", version, revision)

	return 0
}
