package logger

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Level int8

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
	LevelPanic
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	case LevelPanic:
		return "PANIC"
	}
	return ""
}

func ParseLevel(l string) Level {
	switch l {
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN":
		return LevelWarn
	case "ERROR":
		return LevelError
	case "FATAL":
		return LevelFatal
	case "PANIC":
		return LevelPanic
	}
	return LevelInfo
}

type Logger struct {
	newLogger *log.Logger
	level     Level
}

func (l *Logger) Output(level Level, message string) {
	if l.level < level {
		return
	}
	pc, fileName, _, ok := runtime.Caller(2)
	if !ok {
		fileName = "???"
	}
	name := runtime.FuncForPC(pc).Name()
	name = strings.Split(filepath.Base(name), ".")[0]
	message = fmt.Sprintf("%s %s %s [%s] %s",
		time.Now().Format(logTimeFormat), name,
		filepath.Base(fileName), level, message)
	switch level {
	case LevelDebug:
		l.newLogger.Println(message)
	case LevelInfo:
		l.newLogger.Println(message)
	case LevelWarn:
		l.newLogger.Println(message)
	case LevelError:
		l.newLogger.Println(message)
	case LevelFatal:
		l.newLogger.Fatalln(message)
	case LevelPanic:
		l.newLogger.Panicln(message)
	}
}

func (l *Logger) Debug(v ...interface{}) {
	l.Output(LevelDebug, fmt.Sprint(v...))
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.Output(LevelDebug, fmt.Sprintf(format, v...))
}

func (l *Logger) Info(v ...interface{}) {
	l.Output(LevelInfo, fmt.Sprint(v...))
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.Output(LevelInfo, fmt.Sprintf(format, v...))
}

func (l *Logger) Warn(v ...interface{}) {
	l.Output(LevelWarn, fmt.Sprint(v...))
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.Output(LevelWarn, fmt.Sprintf(format, v...))
}

func (l *Logger) Error(v ...interface{}) {
	l.Output(LevelError, fmt.Sprint(v...))
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.Output(LevelError, fmt.Sprintf(format, v...))
}

func (l *Logger) Fatal(v ...interface{}) {
	l.Output(LevelFatal, fmt.Sprint(v...))
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.Output(LevelFatal, fmt.Sprintf(format, v...))
}

func (l *Logger) Panic(v ...interface{}) {
	l.Output(LevelPanic, fmt.Sprint(v...))
}

func (l *Logger) Panicf(format string, v ...interface{}) {
	l.Output(LevelPanic, fmt.Sprintf(format, v...))
}
