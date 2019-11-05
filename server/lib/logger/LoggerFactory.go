package logger

type LoggerFactory interface {
	New(depth ...int) Logger
}

type internalLoggerFactory struct{}

var Factory LoggerFactory = (*internalLoggerFactory)(nil)

func (f *internalLoggerFactory) New(depth ...int) Logger {
	return NewStdLogger(depth...)
}
