// Package poll provides a poller that retrieves financial data for the given
// stock symbols at the given interval and updates the archiver with the
// results.
package poll

import (
	"context"
	"time"

	"github.com/awoodbeck/faang-stonks/finance"
	"github.com/awoodbeck/faang-stonks/history"
	"go.uber.org/zap"
)

// Poller knows how to continually retrieve stock quotes from a
// finance.Provider and store the quotes in a history.Archiver.
type Poller struct {
	log      *zap.SugaredLogger
	archiver history.Archiver
	provider finance.Provider
}

// Poll accepts an interval and a slice of stock symbols, and polls their
// current price at regular intervals, archiving the results.
func (p Poller) Poll(ctx context.Context, interval time.Duration,
	symbols ...string) {
	if len(symbols) == 0 {
		p.log.Warn("no symbols to poll")
		return
	}
	if interval <= 0 {
		p.log.Warn("invalid interval; using default 1 hour")
		interval = 60 * time.Minute
	}

	t := time.NewTicker(interval)
	defer t.Stop()

	for {
		quotes, err := p.provider.GetQuotes(ctx, symbols...)
		if err != nil {
			p.log.Errorf("polling provider: %v", err)
		} else {
			err = p.archiver.SetQuotes(ctx, quotes)
			if err != nil {
				p.log.Errorf("updating history: %v", err)
			}
		}

		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

// New accepts a finance.Provider, a history.Archiver, and a logger, and
// returns a pointer to a poller Client.
func New(p finance.Provider, a history.Archiver, l *zap.SugaredLogger) (
	*Poller, error) {
	return &Poller{
		log:      l,
		archiver: a,
		provider: p,
	}, nil
}
