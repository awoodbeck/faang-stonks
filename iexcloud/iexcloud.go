// Package iexcloud provides a finance.Provider implementation that retrieves
// its data from IEX Cloud's API.
package iexcloud

import (
	"context"

	"github.com/awoodbeck/faang-stonks/finance"
)

var _ finance.Provider = (*Client)(nil)

type Client struct{}

func (c Client) GetQuotes(ctx context.Context, symbol ...string) ([]finance.Quote, error) {
	panic("implement me")
}
