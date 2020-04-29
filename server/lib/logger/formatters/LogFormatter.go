package formatters

type LogFormatter interface {
	Format(prefix string, message string) (string, string)
}
