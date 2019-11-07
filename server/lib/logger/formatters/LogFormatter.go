package formatters

type LogFormatter interface {
	Format(string) string
}
