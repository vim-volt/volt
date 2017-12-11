package logger

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

type LogLevel int

const (
	ErrorLevel LogLevel = 1
	WarnLevel  LogLevel = 2
	InfoLevel  LogLevel = 3
	DebugLevel LogLevel = 4
)

var logLevel = InfoLevel

func Errorf(format string, msgs ...interface{}) {
	if logLevel < ErrorLevel {
		return
	}
	msgs = append([]interface{}{getDebugPrefix()}, msgs...)
	fmt.Fprintf(os.Stderr, "[ERROR]%s "+format+"\n", msgs...)
}

func Error(msgs ...interface{}) {
	if logLevel < ErrorLevel {
		return
	}
	cmsg := getDebugPrefix()
	msgs = append([]interface{}{"[ERROR]" + cmsg}, msgs...)
	fmt.Fprintln(os.Stderr, msgs...)
}

func Warnf(format string, msgs ...interface{}) {
	if logLevel < WarnLevel {
		return
	}
	msgs = append([]interface{}{getDebugPrefix()}, msgs...)
	fmt.Printf("[WARN]%s "+format+"\n", msgs...)
}

func Warn(msgs ...interface{}) {
	if logLevel < WarnLevel {
		return
	}
	cmsg := getDebugPrefix()
	msgs = append([]interface{}{"[WARN]" + cmsg}, msgs...)
	fmt.Println(msgs...)
}

func Infof(format string, msgs ...interface{}) {
	if logLevel < InfoLevel {
		return
	}
	msgs = append([]interface{}{getDebugPrefix()}, msgs...)
	fmt.Printf("[INFO]%s "+format+"\n", msgs...)
}

func Info(msgs ...interface{}) {
	if logLevel < InfoLevel {
		return
	}
	cmsg := getDebugPrefix()
	msgs = append([]interface{}{"[INFO]" + cmsg}, msgs...)
	fmt.Println(msgs...)
}

func Debugf(format string, msgs ...interface{}) {
	if logLevel < DebugLevel {
		return
	}
	msgs = append([]interface{}{getDebugPrefix()}, msgs...)
	fmt.Printf("[DEBUG]%s "+format+"\n", msgs...)
}

func Debug(msgs ...interface{}) {
	if logLevel < DebugLevel {
		return
	}
	cmsg := getDebugPrefix()
	msgs = append([]interface{}{"[DEBUG]" + cmsg}, msgs...)
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
