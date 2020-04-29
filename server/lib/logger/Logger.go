package logger

import (
	"fmt"
	"gbf-proxy/lib/logger/formatters"
	"gbf-proxy/lib/logger/printers"
	"os"
	"strings"
)

type Logger struct {
	Printers   []printers.LogPrinter
	Formatters []formatters.LogFormatter
}

const (
	DEBUG = "DEBUG"
	INFO  = "INFO"
	WARN  = "WARN"
	ERROR = "ERROR"
	FATAL = "FATAL"
)

var DefaultPrinters = []printers.LogPrinter{
	printers.NewStdPrinter(),
}
var DefaultFormatters = []formatters.LogFormatter{
	formatters.NewCallerFormatter(),
}
var DefaultLogger = &Logger{
	Printers:   DefaultPrinters,
	Formatters: DefaultFormatters,
}

func (l *Logger) Debug(v ...interface{}) {
	l.Stdout(l.Sprintln(DEBUG, v...))
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.Stdout(l.Sprintf(DEBUG, format, v...))
}

func (l *Logger) Info(v ...interface{}) {
	l.Stdout(l.Sprintln(INFO, v...))
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.Stdout(l.Sprintf(INFO, format, v...))
}

func (l *Logger) Warn(v ...interface{}) {
	l.Stderr(l.Sprintln(WARN, v...))
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.Stderr(l.Sprintf(WARN, format, v...))
}

func (l *Logger) Error(v ...interface{}) {
	l.Stderr(l.Sprintln(ERROR, v...))
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.Stderr(l.Sprintf(ERROR, format, v...))
}

func (l *Logger) Fatal(v ...interface{}) {
	l.Stderr(l.Sprintln(FATAL, v...))
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.Stderr(l.Sprintf(FATAL, format, v...))
	os.Exit(1)
}

func (l *Logger) Stdout(message string) {
	for _, p := range l.Printers {
		p.Stdout(message)
	}
}

func (l *Logger) Stderr(message string) {
	for _, p := range l.Printers {
		p.Stderr(message)
	}
}

func (l *Logger) Sprintln(level string, v ...interface{}) string {
	message := l.Format(level, fmt.Sprintln(v...))
	return strings.TrimSpace(message)
}

func (l *Logger) Sprintf(level string, format string, v ...interface{}) string {
	return l.Format(level, fmt.Sprintf(format, v...))
}

func (l *Logger) Format(level string, message string) string {
	prefix := ""
	for _, f := range l.Formatters {
		prefix, message = f.Format(prefix, message)
	}
	prefix = strings.TrimSpace(prefix)
	return fmt.Sprintf("[%-5s] %s %s", level, prefix, message)
}
