package printers

type LogPrinter interface {
	Stdout(string)
	Stderr(string)
	Fatal(string)
}
