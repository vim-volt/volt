package cmd

import "fmt"

var version string = "HEAD"

func Version(args []string) int {
	fmt.Println("volt " + version)

	return 0
}
