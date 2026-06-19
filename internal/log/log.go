package log

import (
	"fmt"
	"os"
	"time"
)

var (
	verbose bool
)

func SetVerbose(v bool) {
	verbose = v
}

func Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stdout, "[INFO] %s %s\n", time.Now().Format("15:04:05"), msg)
}

func Warn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stdout, "[WARN] %s %s\n", time.Now().Format("15:04:05"), msg)
}

func Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "[ERROR] %s %s\n", time.Now().Format("15:04:05"), msg)
}

func Debug(format string, args ...interface{}) {
	if !verbose {
		return
	}
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stdout, "[DEBUG] %s %s\n", time.Now().Format("15:04:05"), msg)
}

func Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stdout, "[OK] %s %s\n", time.Now().Format("15:04:05"), msg)
}
