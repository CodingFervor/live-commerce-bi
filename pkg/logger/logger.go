package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Logger struct {
	level  string
	info   *log.Logger
	warn   *log.Logger
	err    *log.Logger
	debug  *log.Logger
}

var std *Logger

func Init(level string) {
	std = &Logger{
		level: level,
		info:  log.New(os.Stdout, "[INFO]  ", log.Ldate|log.Ltime),
		warn:  log.New(os.Stdout, "[WARN]  ", log.Ldate|log.Ltime),
		err:   log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime),
		debug: log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime),
	}
}

func Info(format string, args ...interface{})  { std.info.Output(2, fmt.Sprintf(format, args...)) }
func Warn(format string, args ...interface{})  { std.warn.Output(2, fmt.Sprintf(format, args...)) }
func Error(format string, args ...interface{}) { std.err.Output(2, fmt.Sprintf(format, args...)) }
func Debug(format string, args ...interface{}) {
	if std.level == "debug" {
		std.debug.Output(2, fmt.Sprintf(format, args...))
	}
}

func Since(start time.Time) string {
	return fmt.Sprintf("duration=%s", time.Since(start).Round(time.Millisecond))
}
