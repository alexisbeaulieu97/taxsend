package cli

import (
	"context"

	"taxsend/internal/output"
)

type ctxKey string

const printerKey ctxKey = "printer"

func withPrinter(ctx context.Context, p *output.Printer) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, printerKey, p)
}

func getPrinter(ctx context.Context) *output.Printer {
	if ctx == nil {
		return output.New(false)
	}
	if p, ok := ctx.Value(printerKey).(*output.Printer); ok {
		return p
	}
	return output.New(false)
}
