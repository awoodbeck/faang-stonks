// Package finance provides interfaces and types for financial data services.
//
// Consumers use financial data sources to continually retrieve stock quotes
// for a given subset of stock symbols.
package finance

import "context"

// Provider describes an object that returns one quote per given stock symbol.
type Provider interface {
	GetQuotes(ctx context.Context, symbol ...string) ([]Quote, error)
}
