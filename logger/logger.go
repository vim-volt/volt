package logger

import (
	"fmt"
	"os"
	"runtime"
	"strings"
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
	msgs = append(msgs, getCallerMsg())
	fmt.Fprintf(os.Stderr, "[ERROR]%s "+format+"\n", msgs...)
}

func Error(msgs ...interface{}) {
	if logLevel < ErrorLevel {
		return
	}
	cmsg := getCallerMsg()
	msgs = append([]interface{}{"[ERROR]" + cmsg}, msgs...)
	fmt.Fprintln(os.Stderr, msgs...)
}

func Warnf(format string, msgs ...interface{}) {
	if logLevel < WarnLevel {
		return
	}
	msgs = append(msgs, getCallerMsg())
	fmt.Printf("[WARN]%s "+format+"\n", msgs...)
}

func Warn(msgs ...interface{}) {
	if logLevel < WarnLevel {
		return
	}
	cmsg := getCallerMsg()
	msgs = append([]interface{}{"[WARN]" + cmsg}, msgs...)
	fmt.Println(msgs...)
}

func Infof(format string, msgs ...interface{}) {
	if logLevel < InfoLevel {
		return
	}
	msgs = append(msgs, getCallerMsg())
	fmt.Printf("[INFO]%s "+format+"\n", msgs...)
}

func Info(msgs ...interface{}) {
	if logLevel < InfoLevel {
		return
	}
	cmsg := getCallerMsg()
	msgs = append([]interface{}{"[INFO]" + cmsg}, msgs...)
	fmt.Println(msgs...)
}

func Debugf(format string, msgs ...interface{}) {
	if logLevel < DebugLevel {
		return
	}
	msgs = append(msgs, getCallerMsg())
	fmt.Printf("[DEBUG]%s "+format+"\n", msgs...)
}

func Debug(msgs ...interface{}) {
	if logLevel < DebugLevel {
		return
	}
	cmsg := getCallerMsg()
	msgs = append([]interface{}{"[DEBUG]" + cmsg}, msgs...)
	fmt.Println(msgs...)
}

const voltDirName = "github.com/vim-volt/volt"

func getCallerMsg() string {
	if logLevel < DebugLevel {
		return ""
	}
	_, fn, line, _ := runtime.Caller(2)
	idx := strings.Index(fn, voltDirName)
	if idx >= 0 {
		fn = "(volt)" + fn[idx+len(voltDirName):]
	}
	return fmt.Sprintf("[%s:%d]", fn, line)
}

func SetLevel(level LogLevel) {
	logLevel = level
}
