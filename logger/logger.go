package logger

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
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

var out *color.Color
var m sync.Mutex

func init() {
	if !color.NoColor {
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
	out = color.New()
}

var logLevel = InfoLevel

func Errorf(format string, msgs ...interface{}) {
	if logLevel < ErrorLevel {
		return
	}
	m.Lock()
	defer m.Unlock()
	msgs = append([]interface{}{getDebugPrefix()}, msgs...)
	out.Fprintf(colorable.NewColorableStderr(), errorLabel+"%s "+format+"\n", msgs...)
}

func Error(msgs ...interface{}) {
	if logLevel < ErrorLevel {
		return
	}
	m.Lock()
	defer m.Unlock()
	cmsg := getDebugPrefix()
	msgs = append([]interface{}{errorLabel + cmsg}, msgs...)
	out.Fprintln(colorable.NewColorableStderr(), msgs...)
}

func Warnf(format string, msgs ...interface{}) {
	if logLevel < WarnLevel {
		return
	}
	m.Lock()
	defer m.Unlock()
	msgs = append([]interface{}{getDebugPrefix()}, msgs...)
	out.Printf(warnLabel+"%s "+format+"\n", msgs...)
}

func Warn(msgs ...interface{}) {
	if logLevel < WarnLevel {
		return
	}
	m.Lock()
	defer m.Unlock()
	cmsg := getDebugPrefix()
	msgs = append([]interface{}{warnLabel + cmsg}, msgs...)
	out.Println(msgs...)
}

func Infof(format string, msgs ...interface{}) {
	if logLevel < InfoLevel {
		return
	}
	m.Lock()
	defer m.Unlock()
	msgs = append([]interface{}{getDebugPrefix()}, msgs...)
	out.Printf(infoLabel+"%s "+format+"\n", msgs...)
}

func Info(msgs ...interface{}) {
	if logLevel < InfoLevel {
		return
	}
	m.Lock()
	defer m.Unlock()
	cmsg := getDebugPrefix()
	msgs = append([]interface{}{infoLabel + cmsg}, msgs...)
	out.Println(msgs...)
}

func Debugf(format string, msgs ...interface{}) {
	if logLevel < DebugLevel {
		return
	}
	m.Lock()
	defer m.Unlock()
	msgs = append([]interface{}{getDebugPrefix()}, msgs...)
	out.Printf(debugLabel+"%s "+format+"\n", msgs...)
}

func Debug(msgs ...interface{}) {
	if logLevel < DebugLevel {
		return
	}
	m.Lock()
	defer m.Unlock()
	cmsg := getDebugPrefix()
	msgs = append([]interface{}{debugLabel + cmsg}, msgs...)
	out.Println(msgs...)
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
