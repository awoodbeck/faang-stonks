package memory

import (
	"strings"

	"github.com/awoodbeck/faang-stonks/finance"
)

type Option func(*Client)

// Symbols configures the Archiver to track specific stock symbols.
func Symbols(symbols []string) Option {
	s := make([]string, len(symbols))
	copy(s, symbols)

	return func(c *Client) {
		c.quotes = make(map[string][]finance.Quote)
		for _, symbol := range s {
			c.quotes[strings.ToLower(symbol)] = []finance.Quote{}
		}
	}
}
