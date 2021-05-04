// Package history provides interfaces and types for objects that store store
// a retrieve historical financial data.
package history

import (
	"context"
	"fmt"

	"github.com/awoodbeck/faang-stonks/finance"
)

var ErrNotFound = fmt.Errorf("not found")

// Provider describes an object that can retrieve requested stock quotes.
type Provider interface {
	// GetQuotes accepts a context for cancellation support, a stock symbol,
	// and an integer representing the last n quotes to return. It returns
	// the lesser of the last n quotes or the maximum number of archived
	// quotes.
	GetQuotes(ctx context.Context, symbol string, last int) ([]finance.Quote, error)

	// GetQuotesBatch accepts a context for cancellation support, stock
	// symbols, and an integer representing the last n quotes to return for
	// each symbol. It returns the lesser of the last n quotes or the maximum
	// number of archived quotes. The default symbol list is used in the
	// absence of a populated symbols slice.
	GetQuotesBatch(ctx context.Context, symbols []string, last int) (finance.QuoteBatch, error)
}
