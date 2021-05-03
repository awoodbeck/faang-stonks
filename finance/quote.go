package finance

import "time"

var DefaultSymbols = []string{"fb", "amzn", "aapl", "nflx", "goog"}

// Quote represents the snapshot of a stock's price.
type Quote struct {
	Price  float64   `json:"price"`
	Symbol string    `json:"symbol"`
	Time   time.Time `json:"time"`
}
