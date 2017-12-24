package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
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
	cmdList, err := getCmdList()
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

// Return sorted list of command names list
func getCmdList() ([]string, error) {
	out, err := testutil.RunVolt("help")
	if err != nil {
		return nil, err
	}
	outstr := string(out)
	lines := strings.Split(outstr, "\n")
	cmdidx := -1
	for i := range lines {
		if lines[i] == "Command" {
			cmdidx = i + 1
			break
		}
	}
	if cmdidx < 0 {
		return nil, errors.New("not found 'Command' line in 'volt help'")
	}
	dup := make(map[string]bool, 20)
	cmdList := make([]string, 0, 20)
	re := regexp.MustCompile(`^  (\S+)`)
	for i := cmdidx; i < len(lines); i++ {
		if m := re.FindStringSubmatch(lines[i]); len(m) != 0 && !dup[m[1]] {
			cmdList = append(cmdList, m[1])
			dup[m[1]] = true
		}
	}
	sort.Strings(cmdList)
	return cmdList, nil
}
