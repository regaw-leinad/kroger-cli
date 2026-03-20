package output

import (
	"context"
	"os"

	"golang.org/x/term"
)

type ctxKeyPrinter struct{}

// WithPrinter stores the printer in context.
func WithPrinter(ctx context.Context, p Printer) context.Context {
	return context.WithValue(ctx, ctxKeyPrinter{}, p)
}

// From retrieves the printer from context.
func From(ctx context.Context) Printer {
	if p, ok := ctx.Value(ctxKeyPrinter{}).(Printer); ok {
		return p
	}
	return NewHumanPrinter(os.Stdout, os.Stderr)
}

// IsTerminal returns true if stdout is a terminal.
func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
