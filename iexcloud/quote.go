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
	Price     float64 `json:"delayedPrice"`
	Timestamp int64   `json:"delayedPriceTime"`
}

type batch map[string]map[string]quote

func (b batch) MarshalQuotes() ([]finance.Quote, error) {
	quotes := make([]finance.Quote, 0, len(b))

	for symbol := range b {
		q, ok := b[symbol]["quote"]
		if !ok {
			return nil, fmt.Errorf("unable to retrieve quote data for %q", symbol)
		}
		quotes = append(quotes, finance.Quote{
			Symbol: q.Symbol,
			Price:  q.Price,
			Time:   time.Unix(q.Timestamp/1000, q.Timestamp%1000),
		})
	}

	return quotes, nil
}
