// Package ram provides history.Archiver and history.Provider implementations
// that use RAM for volatile storage.
//
// The requirements regarding volatility weren't specific, but this
// implementation may fit the business case when we only care about stock
// prices while the service runs. The downside is RAM is more expensive than
// disk, so we have to be mindful of the growing memory consumption of this
// approach. That said, we could use something like a fixed queue to limit
// the number of quotes per symbol we track.
//
// Side note: I punt on time zone handling in this example, storing timestamps
// as-is. That isn't suitable in production, of course. You can see how I
// approach this problem in the sqlite.Client.
package ram

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/awoodbeck/faang-stonks/finance"
	"github.com/awoodbeck/faang-stonks/history"
	"go.uber.org/multierr"
)

var (
	_ history.Archiver = (*Client)(nil)
	_ history.Provider = (*Client)(nil)
)

// Client implements the history.Archiver and history.Provider interfaces,
// knowing how to store and retrieve stock quotes, respectively.
type Client struct {
	mu     sync.RWMutex
	quotes map[string][]finance.Quote
}

// Close is essentially a no-op for this client.
func (c *Client) Close() error {
	return nil
}

// GetQuotes accepts a stock symbol and the last N quotes for the stock to
// return.
func (c *Client) GetQuotes(_ context.Context, symbol string, last int) (
	[]finance.Quote, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	quotes, ok := c.quotes[strings.ToLower(symbol)]
	if !ok {
		return nil, fmt.Errorf("quotes for %q not found", symbol)
	}

	if last < 1 {
		last = 1
	}
	if len(quotes) < last {
		last = len(quotes)
	}

	out := make([]finance.Quote, last)
	copy(out, quotes)

	return out, nil
}

// SetQuotes accepts a slice of finance.Quote objects and archives them to
// the appropriate in-memory slice.
func (c *Client) SetQuotes(_ context.Context, quotes []finance.Quote) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var err error
	for _, quote := range quotes {
		symbol := strings.ToLower(quote.Symbol)
		quotes, ok := c.quotes[symbol]
		if !ok {
			multierr.AppendInto(&err, fmt.Errorf("symbol %q not found", quote.Symbol))
			continue
		}

		c.quotes[symbol] = append([]finance.Quote{quote}, quotes...)
	}

	return err
}

// New returns a pointer to a new Client object after applying optional settings.
func New(options ...Option) (*Client, error) {
	c := &Client{quotes: make(map[string][]finance.Quote)}

	for _, symbol := range finance.DefaultSymbols {
		c.quotes[strings.ToLower(symbol)] = []finance.Quote{}
	}

	for _, option := range options {
		option(c)
	}

	return c, nil
}
