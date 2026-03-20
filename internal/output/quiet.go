package output

// QuietPrinter wraps another printer but suppresses status/success messages.
type QuietPrinter struct {
	Printer
}

func NewQuietPrinter(inner Printer) *QuietPrinter {
	return &QuietPrinter{Printer: inner}
}

func (p *QuietPrinter) Status(format string, args ...any) {}

func (p *QuietPrinter) Success(format string, args ...any) {}
