package logger

import (
	"fmt"
	"net/http"
	"strings"
)

type RequestLogger struct {
	*http.Request
	Logger
}

var _ Logger = (*RequestLogger)(nil)

func NewRequestLogger(req *http.Request, l Logger) *RequestLogger {
	return &RequestLogger{
		Request: req,
		Logger:  l,
	}
}

func (l *RequestLogger) Debug(v ...interface{}) {
	l.Logger.Debug(l.Sprintln(v...))
}

func (l *RequestLogger) Debugf(msg string, v ...interface{}) {
	l.Logger.Debugf(l.Sprintln(msg), v...)
}

func (l *RequestLogger) Info(v ...interface{}) {
	l.Logger.Info(l.Sprintln(v...))
}

func (l *RequestLogger) Infof(format string, v ...interface{}) {
	l.Logger.Infof(l.Sprintln(format), v...)
}

func (l *RequestLogger) Warn(v ...interface{}) {
	l.Logger.Warn(l.Sprintln(v...))
}

func (l *RequestLogger) Warnf(format string, v ...interface{}) {
	l.Logger.Warnf(l.Sprintln(format), v...)
}

func (l *RequestLogger) Error(v ...interface{}) {
	l.Logger.Error(l.Sprintln(v...))
}

func (l *RequestLogger) Errorf(format string, v ...interface{}) {
	l.Logger.Errorf(l.Sprintln(format), v...)
}

func (l *RequestLogger) Fatal(v ...interface{}) {
	l.Logger.Fatal(l.Sprintln(v...))
}

func (l *RequestLogger) Fatalf(format string, v ...interface{}) {
	l.Logger.Fatalf(l.Sprintln(format), v...)
}

func (l *RequestLogger) Sprintln(v ...interface{}) string {
	forwardedFor := l.Request.Header.Get("X-Forwarded-For")
	return strings.TrimSpace(fmt.Sprintf("[%s] %s", forwardedFor, fmt.Sprintln(v...)))
}
