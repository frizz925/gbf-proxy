package logger

type LoggerFactory interface {
	New(depth ...int) Logger
}

type internalLoggerFactory struct{}

var Factory LoggerFactory = (*internalLoggerFactory)(nil)

var DefaultLogger = Factory.New()

func (f *internalLoggerFactory) New(depth ...int) Logger {
	return NewStdLogger(depth...)
}
