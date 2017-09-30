package logger

import (
	"fmt"
	"os"
)

func Errorf(format string, msgs ...interface{}) {
	fmt.Fprintf(os.Stderr, "[ERROR] "+format+"\n", msgs...)
}

func Error(msgs ...interface{}) {
	msgs = append([]interface{}{"[ERROR]"}, msgs...)
	fmt.Fprintln(os.Stderr, msgs...)
}

func Warnf(format string, msgs ...interface{}) {
	fmt.Printf("[WARN] "+format+"\n", msgs...)
}

func Warn(msgs ...interface{}) {
	msgs = append([]interface{}{"[WARN]"}, msgs...)
	fmt.Println(msgs...)
}

func Infof(format string, msgs ...interface{}) {
	fmt.Printf("[INFO] "+format+"\n", msgs...)
}

func Info(msgs ...interface{}) {
	msgs = append([]interface{}{"[INFO]"}, msgs...)
	fmt.Println(msgs...)
}
