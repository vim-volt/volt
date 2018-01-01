package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/vim-volt/volt/internal/testutil"
)

func main() {
	os.Exit(Main())
}

// Embeds "volt help" output in the first code block (lines surrounded by ```)
// of README.md
func Main() int {
	if len(os.Args) <= 1 {
		fmt.Fprintln(os.Stderr, "[WARN] Specify README.md path")
		return 1
	}
	readme := os.Args[1]
	fileinfo, err := os.Stat(readme)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[WARN] %s\n", err.Error())
		return 2
	}
	file, err := os.OpenFile(os.Args[1], os.O_RDWR, fileinfo.Mode())
	if err != nil {
		fmt.Fprintf(os.Stderr, "[WARN] %s\n", err.Error())
		return 3
	}
	defer file.Close()

	// Read content from file
	var content bytes.Buffer
	if _, err := io.Copy(&content, file); err != nil {
		fmt.Fprintf(os.Stderr, "[WARN] %s\n", err.Error())
		return 4
	}
	// Find the first code block ("volt help" output)
	lines := strings.Split(content.String(), "\n")
	start, end := findTopCodeBlockRange(lines)
	if start < 0 {
		fmt.Fprintln(os.Stderr, "[WARN] Cannot find code block")
		return 5
	}

	// seek for writing to file
	if _, err := file.Seek(0, 0); err != nil {
		fmt.Fprintf(os.Stderr, "[WARN] %s\n", err.Error())
		return 6
	}
	bw := bufio.NewWriter(file)
	for _, line := range lines[:start] {
		bw.WriteString(line + "\n")
	}
	out, err := getVoltHelpOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[WARN] %s\n", err.Error())
		return 7
	}
	bw.WriteString(out + "\n")
	for _, line := range lines[end:] {
		bw.WriteString(line + "\n")
	}

	if err := bw.Flush(); err != nil {
		fmt.Fprintln(os.Stderr, "[WARN] Cannot find code block")
		return 8
	}
	curpos, err := file.Seek(0, 1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[WARN] %s\n", err.Error())
		return 9
	}
	// Specify curpos-1 to delete the last newline
	if err := file.Truncate(curpos - 1); err != nil {
		fmt.Fprintf(os.Stderr, "[WARN] %s\n", err.Error())
		return 10
	}
	return 0
}

func findTopCodeBlockRange(lines []string) (int, int) {
	return 6, 80
}

func getVoltHelpOutput() (string, error) {
	out, err := testutil.RunVolt("help")
	if err != nil {
		return "", err
	}
	return "$ volt\n" + strings.TrimRight(string(out), " \t\r\n"), nil
}
