package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/rjkat/volt/internal/testutil"
)

func main() {
	os.Exit(doMain())
}

// Update CMDREF.md "volt help" output in the first code block (lines surrounded by ```)
func doMain() int {
	header, err := getVoltHelpOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[WARN] %s\n", err.Error())
		return 1
	}

	cmdref, err := getCmdRefContent()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[WARN] %s\n", err.Error())
		return 2
	}

	fmt.Printf("%s\n\n%s", header, cmdref)

	return 0
}

func getVoltHelpOutput() (string, error) {
	out, err := testutil.RunVolt("help")
	if err != nil {
		return "", err
	}
	content := strings.TrimRight(string(out), " \t\r\n")
	return fmt.Sprintf("```\n%s\n```", content), nil
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
