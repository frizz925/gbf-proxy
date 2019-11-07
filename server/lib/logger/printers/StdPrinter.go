package printers

import (
	"fmt"
	"log"
	"os"
)

type StdPrinter struct {
	Logger    *log.Logger
	ErrLogger *log.Logger
}

var _ LogPrinter = (*StdPrinter)(nil)

func NewStdPrinter() *StdPrinter {
	flags := log.LstdFlags | log.LUTC
	return &StdPrinter{
		Logger:    log.New(os.Stdout, "", flags),
		ErrLogger: log.New(os.Stderr, "", flags),
	}
}

func (p *StdPrinter) Stdout(text string) {
	p.Logger.Println(text)
}

func (p *StdPrinter) Stderr(text string) {
	p.ErrLogger.Println(text)
}

func (p *StdPrinter) Fatal(text string) {
	p.ErrLogger.Fatal(text)
}

func (*StdPrinter) Format(level string, message string) string {
	return fmt.Sprintf("[%s] %s", level, message)
}
