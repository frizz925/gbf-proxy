package logging

import (
	"fmt"
	"io"
	"log"
	"os"
)

type Logger struct {
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

func New(config *LoggerConfig) *Logger {
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

	return &Logger{
		Name:      config.Name,
		Logger:    logger,
		ErrLogger: errLogger,
	}
}

func (l *Logger) Debugf(format string, a ...interface{}) {
	l.Debug(fmt.Sprintf(format, a...))
}

func (l *Logger) Debug(a ...interface{}) {
	l.Log("debug", a...)
}

func (l *Logger) Infof(format string, a ...interface{}) {
	l.Info(fmt.Sprintf(format, a...))
}

func (l *Logger) Info(a ...interface{}) {
	l.Log("info", a...)
}

func (l *Logger) Warnf(format string, a ...interface{}) {
	l.Warn(fmt.Sprintf(format, a...))
}

func (l *Logger) Warn(a ...interface{}) {
	l.LogErr("warn", a...)
}

func (l *Logger) Errorf(format string, a ...interface{}) {
	l.Error(fmt.Sprintf(format, a...))
}

func (l *Logger) Error(a ...interface{}) {
	l.LogErr("error", a...)
}

func (l *Logger) Log(level string, a ...interface{}) {
	l.Logger.Println(l.Format(level, a...))
}

func (l *Logger) LogErr(level string, a ...interface{}) {
	l.ErrLogger.Println(l.Format(level, a...))
}

func (l *Logger) Format(level string, a ...interface{}) string {
	message := fmt.Sprint(a...)
	return fmt.Sprintf("[%s] [%s] %s", l.Name, level, message)
}
