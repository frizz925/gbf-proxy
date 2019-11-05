package logger

import (
	"fmt"
	"net"
)

type ConnectionLogger struct {
	net.Conn
	Logger
}

var _ Logger = (*ConnectionLogger)(nil)

func NewConnectionLogger(conn net.Conn, l Logger) *ConnectionLogger {
	return &ConnectionLogger{
		Conn:   conn,
		Logger: l,
	}
}

func (l *ConnectionLogger) Debug(v ...interface{}) {
	l.Logger.Debug(l.Sprintln(v...))
}

func (l *ConnectionLogger) Debugf(msg string, v ...interface{}) {
	l.Logger.Debugf(l.Sprintln(msg), v...)
}

func (l *ConnectionLogger) Info(v ...interface{}) {
	l.Logger.Info(l.Sprintln(v...))
}

func (l *ConnectionLogger) Infof(format string, v ...interface{}) {
	l.Logger.Infof(l.Sprintln(format), v...)
}

func (l *ConnectionLogger) Warn(v ...interface{}) {
	l.Logger.Warn(l.Sprintln(v...))
}

func (l *ConnectionLogger) Warnf(format string, v ...interface{}) {
	l.Logger.Warnf(l.Sprintln(format), v...)
}

func (l *ConnectionLogger) Error(v ...interface{}) {
	l.Logger.Error(l.Sprintln(v...))
}

func (l *ConnectionLogger) Errorf(format string, v ...interface{}) {
	l.Logger.Errorf(l.Sprintln(format), v...)
}

func (l *ConnectionLogger) Fatal(v ...interface{}) {
	l.Logger.Fatal(l.Sprintln(v...))
}

func (l *ConnectionLogger) Fatalf(format string, v ...interface{}) {
	l.Logger.Fatalf(l.Sprintln(format), v...)
}

func (l *ConnectionLogger) Sprintln(v ...interface{}) string {
	remoteAddr := l.Conn.RemoteAddr().String()
	return fmt.Sprintf("[%s] %s", remoteAddr, fmt.Sprintln(v...))
}
