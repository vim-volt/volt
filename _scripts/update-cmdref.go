package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/vim-volt/volt/internal/testutil"
)

func main() {
	os.Exit(Main())
}

// Update CMDREF.md "volt help" output in the first code block (lines surrounded by ```)
func Main() int {
	if len(os.Args) <= 1 {
		fmt.Fprintln(os.Stderr, "[WARN] Specify CMDREF.md path")
		return 1
	}
	file, err := os.Create(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "[WARN] %s\n", err.Error())
		return 2
	}
	defer file.Close()

	bw := bufio.NewWriter(file)
	out, err := getCmdRefContent()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[WARN] %s\n", err.Error())
		return 3
	}
	bw.WriteString(out)

	if err := bw.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "[WARN] %s\n", err.Error())
		return 4
	}
	return 0
}

func getCmdRefContent() (string, error) {
	cmdList, err := testutil.GetCmdList()
	if err != nil {
		return "", err
	}
	sections := make([]string, 0, len(cmdList))
	for _, cmd := range cmdList {
		out, err := testutil.RunVolt("help", cmd)
		if err != nil {
			return "", errors.New("volt help " + cmd + ": " + err.Error())
		}
		outstr := string(out)
		outstr = strings.Trim(outstr, " \t\r\n")
		outstr = strings.Replace(outstr, "\t", "    ", -1)
		s := fmt.Sprintf("# volt %s\n\n```\n%s\n```", cmd, outstr)
		sections = append(sections, s)
	}
	return strings.Join(sections, "\n\n"), nil
}
