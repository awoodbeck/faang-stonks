// Package sqlite provides history.Archiver and history.Provider
// implementations that use SQLite for storage.
package sqlite

import (
	"context"
	"database/sql"

	"github.com/awoodbeck/faang-stonks/finance"
	"github.com/awoodbeck/faang-stonks/history"
)

var (
	_ history.Archiver = (*Client)(nil)
	_ history.Provider = (*Client)(nil)
)

type Client struct {
	db *sql.DB
}

func (c Client) GetQuotes(ctx context.Context, symbol string, count int) ([]finance.Quote, error) {
	panic("implement me")
}

func (c Client) SetQuotes(ctx context.Context, quotes []finance.Quote) error {
	panic("implement me")
}
