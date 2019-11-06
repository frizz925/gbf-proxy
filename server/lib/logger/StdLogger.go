package logger

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
)

type StdLogger struct {
	Logger                *log.Logger
	ErrLogger             *log.Logger
	ReportingDepth        int
	InitialReportingDepth int
}

var _ Logger = (*StdLogger)(nil)

func NewStdLogger(depth ...int) *StdLogger {
	reportingDepth := 0
	if len(depth) > 0 {
		reportingDepth = depth[0]
	}
	flags := log.LstdFlags | log.LUTC
	return &StdLogger{
		Logger:                log.New(os.Stdout, "", flags),
		ErrLogger:             log.New(os.Stderr, "", flags),
		ReportingDepth:        reportingDepth,
		InitialReportingDepth: 3,
	}
}

func (l *StdLogger) SetReportingDepth(depth int) {
	l.ReportingDepth = depth
}

func (l *StdLogger) Debug(v ...interface{}) {
	l.Stdout(DEBUG, l.Sprintln(v...))
}

func (l *StdLogger) Debugf(format string, v ...interface{}) {
	l.Stdout(DEBUG, l.Sprintf(format, v...))
}

func (l *StdLogger) Info(v ...interface{}) {
	l.Stdout(INFO, l.Sprintln(v...))
}

func (l *StdLogger) Infof(format string, v ...interface{}) {
	l.Stdout(INFO, l.Sprintf(format, v...))
}

func (l *StdLogger) Warn(v ...interface{}) {
	l.Stderr(WARN, l.Sprintln(v...))
}

func (l *StdLogger) Warnf(format string, v ...interface{}) {
	l.Stderr(WARN, l.Sprintf(format, v...))
}

func (l *StdLogger) Error(v ...interface{}) {
	l.Stderr(ERROR, l.Sprintln(v...))
}

func (l *StdLogger) Errorf(format string, v ...interface{}) {
	l.Stderr(ERROR, l.Sprintf(format, v...))
}

func (l *StdLogger) Fatal(v ...interface{}) {
	l.Stderr(FATAL, l.Sprintln(v...))
}

func (l *StdLogger) Fatalf(format string, v ...interface{}) {
	l.Stderr(FATAL, l.Sprintf(format, v...))
	os.Exit(1)
}

func (l *StdLogger) Stdout(level string, msg string) {
	l.Printf(l.Logger, level, msg)
}

func (l *StdLogger) Stderr(level string, msg string) {
	l.Printf(l.ErrLogger, level, msg)
}

func (l *StdLogger) Printf(ll *log.Logger, level string, msg string) {
	ll.Printf("[%s] %s", level, msg)
}

func (l *StdLogger) Sprintln(v ...interface{}) string {
	return fmt.Sprintf("[%s] %s", l.GetSource(), fmt.Sprintln(v...))
}

func (l *StdLogger) Sprintf(format string, v ...interface{}) string {
	return fmt.Sprintf("[%s] %s", l.GetSource(), fmt.Sprintf(format, v...))
}

func (l *StdLogger) GetSource() string {
	depth := l.InitialReportingDepth + l.ReportingDepth
	_, file, _, _ := runtime.Caller(depth)
	base := path.Base(file)
	idx := strings.LastIndex(base, ".")
	if idx >= 0 {
		return base[:idx]
	}
	return base
}
