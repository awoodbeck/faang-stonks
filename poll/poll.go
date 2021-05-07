// Package poll provides a poller that retrieves financial data for the given
// stock symbols at the given interval and updates the archiver with the
// results.
package poll

import (
	"context"
	"fmt"
	"time"

	"github.com/awoodbeck/faang-stonks/finance"
	"github.com/awoodbeck/faang-stonks/history"
	"go.uber.org/zap"
)

// DefaultPollDuration is the default duration between stock quote updates.
const DefaultPollDuration = time.Minute

var (
	ErrNilArchiver = fmt.Errorf("archiver cannot be nil")
	ErrNilLogger   = fmt.Errorf("logger cannot be nil")
	ErrNilProvider = fmt.Errorf("finance provider cannot be nil")
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
		p.log.Warn("invalid interval; using default 1 minute")
		interval = DefaultPollDuration
	}

	t := time.NewTicker(interval)
	defer t.Stop()

	for {
		quotes, err := p.provider.GetQuotes(ctx, symbols...)
		if err != nil {
			p.log.Errorf("polling provider: %v", err)
		} else {
			p.log.Debugf("received: %#v", quotes)
			err = p.archiver.SetQuotes(ctx, quotes)
			if err != nil {
				p.log.Errorf("updating history: %v", err)
				continue
			}
			p.log.Debug("stored")
		}

		select {
		case <-ctx.Done():
			p.log.Debug("stopping poller")
			return
		case <-t.C:
		}
	}
}

// New accepts a finance.Provider, a history.Archiver, and a logger, and
// returns a pointer to a poller Client.
func New(p finance.Provider, a history.Archiver, l *zap.SugaredLogger) (
	*Poller, error) {
	switch {
	case p == nil:
		return nil, ErrNilProvider
	case a == nil:
		return nil, ErrNilArchiver
	case l == nil:
		return nil, ErrNilLogger
	}

	return &Poller{
		log:      l.Named("poll"),
		archiver: a,
		provider: p,
	}, nil
}
