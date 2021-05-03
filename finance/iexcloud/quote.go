package iexcloud

import (
	"fmt"
	"time"

	"github.com/awoodbeck/faang-stonks/finance"
)

// quote is an IEX Cloud-specific quote. It's used as an intermediate type to
// translate an IEX Cloud JSON to the finance.Quote type.
type quote struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"latestPrice"`
	Timestamp int64   `json:"latestUpdate"`
}

type batchQuotes map[string]map[string]quote

// MarshalQuotes returns a slice of quotes suitable for use in the finance
// package.
func (b batchQuotes) MarshalQuotes() ([]finance.Quote, error) {
	quotes := make([]finance.Quote, 0, len(b))

	for symbol := range b {
		q, ok := b[symbol]["quote"]
		if !ok {
			return nil, fmt.Errorf("'quote' key for symbol '%s' not found", symbol)
		}
		quotes = append(quotes, finance.Quote{
			Symbol: q.Symbol,
			Price:  q.Price,
			Time:   time.Unix(q.Timestamp/1000, q.Timestamp%1000),
		})
	}

	return quotes, nil
}
