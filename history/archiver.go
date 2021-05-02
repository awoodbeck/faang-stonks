package history

import (
	"context"
	"io"

	"github.com/awoodbeck/faang-stonks/finance"
)

// Archiver describes an object that can archive or store stock prices.
type Archiver interface {
	// SetQuotes accepts a slice of finance.Quote objects and archives them.
	SetQuotes(ctx context.Context, quotes []finance.Quote) error
	io.Closer
}
