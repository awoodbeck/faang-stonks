// Package history provides interfaces and types for objects that store store
// a retrieve historical financial data.
package history

import (
	"context"

	"github.com/awoodbeck/faang-stonks/finance"
)

// Archiver describes an object that can archive or store stock prices.
type Archiver interface {
	SetQuotes(ctx context.Context, quotes []finance.Quote) error
}

// Provider describes an object that can retrieve requested stock quotes.
type Provider interface {
	// GetQuotes accepts a context for cancellation support, a stock symbol,
	// and a integer representing the number of quotes to return. It returns
	// the lesser of the requested number of quotes or the maximum number of
	// archived quotes.
	GetQuotes(ctx context.Context, symbol string, count int) ([]finance.Quote, error)
}
