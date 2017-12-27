package logger

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

type LogLevel int

const (
	ErrorLevel LogLevel = 1
	WarnLevel  LogLevel = 2
	InfoLevel  LogLevel = 3
	DebugLevel LogLevel = 4
)

var (
	errorLabel string
	warnLabel  string
	infoLabel  string
	debugLabel string
)

func init() {
	if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		errorLabel = "[" + color.New(color.FgRed).Sprint("ERROR") + "]"
		warnLabel = "[" + color.New(color.FgYellow).Sprint("WARN") + "]"
		infoLabel = "[" + color.New(color.FgCyan).Sprint("INFO") + "]"
		debugLabel = "[" + color.New(color.FgMagenta).Sprint("DEBUG") + "]"
	} else {
		errorLabel = "[ERROR]"
		warnLabel = "[WARN]"
		infoLabel = "[INFO]"
		debugLabel = "[DEBUG]"
	}
}

var logLevel = InfoLevel

func Errorf(format string, msgs ...interface{}) {
	if logLevel < ErrorLevel {
		return
	}
	msgs = append([]interface{}{getDebugPrefix()}, msgs...)
	fmt.Fprintf(os.Stderr, errorLabel+"%s "+format+"\n", msgs...)
}

func Error(msgs ...interface{}) {
	if logLevel < ErrorLevel {
		return
	}
	cmsg := getDebugPrefix()
	msgs = append([]interface{}{errorLabel + cmsg}, msgs...)
	fmt.Fprintln(os.Stderr, msgs...)
}

func Warnf(format string, msgs ...interface{}) {
	if logLevel < WarnLevel {
		return
	}
	msgs = append([]interface{}{getDebugPrefix()}, msgs...)
	fmt.Printf(warnLabel+"%s "+format+"\n", msgs...)
}

func Warn(msgs ...interface{}) {
	if logLevel < WarnLevel {
		return
	}
	cmsg := getDebugPrefix()
	msgs = append([]interface{}{warnLabel + cmsg}, msgs...)
	fmt.Println(msgs...)
}

func Infof(format string, msgs ...interface{}) {
	if logLevel < InfoLevel {
		return
	}
	msgs = append([]interface{}{getDebugPrefix()}, msgs...)
	fmt.Printf(infoLabel+"%s "+format+"\n", msgs...)
}

func Info(msgs ...interface{}) {
	if logLevel < InfoLevel {
		return
	}
	cmsg := getDebugPrefix()
	msgs = append([]interface{}{infoLabel + cmsg}, msgs...)
	fmt.Println(msgs...)
}

func Debugf(format string, msgs ...interface{}) {
	if logLevel < DebugLevel {
		return
	}
	msgs = append([]interface{}{getDebugPrefix()}, msgs...)
	fmt.Printf(debugLabel+"%s "+format+"\n", msgs...)
}

func Debug(msgs ...interface{}) {
	if logLevel < DebugLevel {
		return
	}
	cmsg := getDebugPrefix()
	msgs = append([]interface{}{debugLabel + cmsg}, msgs...)
	fmt.Println(msgs...)
}

func getDebugPrefix() string {
	const voltDirName = "github.com/vim-volt/volt/"
	if logLevel < DebugLevel {
		return ""
	}
	_, fn, line, _ := runtime.Caller(2)
	idx := strings.Index(fn, voltDirName)
	if idx >= 0 {
		fn = fn[idx+len(voltDirName):]
	}
	return fmt.Sprintf("[%s][%s:%d]", time.Now().UTC().Format("15:04:05.000"), fn, line)
}

func SetLevel(level LogLevel) {
	logLevel = level
}
