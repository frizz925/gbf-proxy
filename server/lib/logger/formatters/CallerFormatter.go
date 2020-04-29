package formatters

import (
	"fmt"
	"path"
	"runtime"
	"strings"
	"sync"
)

type CallerFormatter struct {
	callerDepth int
	once        *sync.Once
}

var _ LogFormatter = (*CallerFormatter)(nil)

var (
	minimumCallerDepth = 4
	maximumCallerDepth = 25

	packageCallerDepth = 10
	packageNamespace   = "gbf-proxy/lib/logger"
)

func NewCallerFormatter() *CallerFormatter {
	return &CallerFormatter{
		callerDepth: minimumCallerDepth,
		once:        &sync.Once{},
	}
}

func (f *CallerFormatter) Format(prefix string, message string) (string, string) {
	return fmt.Sprintf("%s [%s]", prefix, f.getSource()), message
}

func (f *CallerFormatter) getSource() string {
	f.once.Do(func() {
		pcs := make([]uintptr, packageCallerDepth)
		depth := runtime.Callers(minimumCallerDepth, pcs)
		frames := runtime.CallersFrames(pcs[:depth])
		callerDepth := 0
		for f, ok := frames.Next(); ok; f, ok = frames.Next() {
			if !strings.HasPrefix(f.Function, packageNamespace) {
				break
			}
			callerDepth++
		}
		f.callerDepth = callerDepth
	})
	_, file, _, _ := runtime.Caller(f.callerDepth)
	base := path.Base(file)
	idx := strings.Index(base, ".")
	if idx >= 0 {
		return base[:idx]
	}
	return base
}
