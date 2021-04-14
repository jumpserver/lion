package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"

	"lion/pkg/config"
)

var globalLogger = &Logger{newLogger: log.New(os.Stdout, "", log.Lmsgprefix)}

const logTimeFormat = "2006-01-02 15:04:05"

func SetupLogger(conf *config.Config) {
	fileName := filepath.Join(conf.LogDirPath, "guacamole.log")
	loggerWriter := &lumberjack.Logger{
		Filename:  fileName,
		MaxSize:   5,
		MaxAge:    7,
		LocalTime: true,
		Compress:  true,
	}
	writer := io.MultiWriter(loggerWriter, os.Stdout)
	l := log.New(writer, "", log.Lmsgprefix)
	globalLogger = &Logger{newLogger: l, level: ParseLevel(conf.LogLevel)}
}

func Debug(v ...interface{}) {
	globalLogger.Output(LevelDebug, fmt.Sprint(v...))
}

func Debugf(format string, v ...interface{}) {
	globalLogger.Output(LevelDebug, fmt.Sprintf(format, v...))
}

func Info(v ...interface{}) {
	globalLogger.Output(LevelInfo, fmt.Sprint(v...))
}

func Infof(format string, v ...interface{}) {
	globalLogger.Output(LevelInfo, fmt.Sprintf(format, v...))
}

func Warn(v ...interface{}) {
	globalLogger.Output(LevelWarn, fmt.Sprint(v...))
}

func Warnf(format string, v ...interface{}) {
	globalLogger.Output(LevelWarn, fmt.Sprintf(format, v...))
}

func Error(v ...interface{}) {
	globalLogger.Output(LevelError, fmt.Sprint(v...))
}

func Errorf(format string, v ...interface{}) {
	globalLogger.Output(LevelError, fmt.Sprintf(format, v...))
}

func Fatal(v ...interface{}) {
	globalLogger.Output(LevelFatal, fmt.Sprint(v...))
}

func Fatalf(format string, v ...interface{}) {
	globalLogger.Output(LevelFatal, fmt.Sprintf(format, v...))
}

func Panic(v ...interface{}) {
	globalLogger.Output(LevelPanic, fmt.Sprint(v...))
}

func Panicf(format string, v ...interface{}) {
	globalLogger.Output(LevelPanic, fmt.Sprintf(format, v...))
}
