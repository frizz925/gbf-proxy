package logging

import (
	"fmt"
	"io"
	"log"
	"os"
)

type Logger interface {
	Debugf(format string, a ...interface{})
	Debug(a ...interface{})
	Infof(format string, a ...interface{})
	Info(a ...interface{})
	Warnf(format string, a ...interface{})
	Warn(a ...interface{})
	Errorf(format string, a ...interface{})
	Error(a ...interface{})
}

type LoggerStd struct {
	Name      string
	Logger    *log.Logger
	ErrLogger *log.Logger
}

type LoggerConfig struct {
	Name      string
	Writer    io.Writer
	ErrWriter io.Writer
}

type nullWriter struct{}

var DefaultWriter io.Writer = os.Stdout
var DefaultErrWriter io.Writer = os.Stderr
var NullWriter = &nullWriter{}

func (w *nullWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

func New(config *LoggerConfig) Logger {
	writer := config.Writer
	if writer == nil {
		writer = DefaultWriter
	}
	logger := log.New(writer, "", log.LstdFlags)

	errWriter := config.ErrWriter
	if errWriter == nil {
		errWriter = DefaultErrWriter
	}
	errLogger := log.New(errWriter, "", log.LstdFlags)

	return &LoggerStd{
		Name:      config.Name,
		Logger:    logger,
		ErrLogger: errLogger,
	}
}

func (l *LoggerStd) Debugf(format string, a ...interface{}) {
	l.Debug(fmt.Sprintf(format, a...))
}

func (l *LoggerStd) Debug(a ...interface{}) {
	l.Log("debug", a...)
}

func (l *LoggerStd) Infof(format string, a ...interface{}) {
	l.Info(fmt.Sprintf(format, a...))
}

func (l *LoggerStd) Info(a ...interface{}) {
	l.Log("info", a...)
}

func (l *LoggerStd) Warnf(format string, a ...interface{}) {
	l.Warn(fmt.Sprintf(format, a...))
}

func (l *LoggerStd) Warn(a ...interface{}) {
	l.LogErr("warn", a...)
}

func (l *LoggerStd) Errorf(format string, a ...interface{}) {
	l.Error(fmt.Sprintf(format, a...))
}

func (l *LoggerStd) Error(a ...interface{}) {
	l.LogErr("error", a...)
}

func (l *LoggerStd) Log(level string, a ...interface{}) {
	l.Logger.Println(l.Format(level, a...))
}

func (l *LoggerStd) LogErr(level string, a ...interface{}) {
	l.ErrLogger.Println(l.Format(level, a...))
}

func (l *LoggerStd) Format(level string, a ...interface{}) string {
	message := fmt.Sprint(a...)
	return fmt.Sprintf("[%s] [%s] %s", l.Name, level, message)
}
