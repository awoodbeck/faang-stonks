// Package influxdb provides history.Archiver and history.Provider
// implementations that use the time series database, InfluxDB.
package influxdb

import (
	"context"

	"github.com/awoodbeck/faang-stonks/finance"
	"github.com/awoodbeck/faang-stonks/history"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

var (
	_ history.Archiver = (*Client)(nil)
	_ history.Provider = (*Client)(nil)
)

type Client struct {
	idb influxdb2.Client
}

func (c Client) GetQuotes(ctx context.Context, symbol string, count int) ([]finance.Quote, error) {
	panic("implement me")
}

func (c Client) SetQuotes(ctx context.Context, quotes []finance.Quote) error {
	panic("implement me")
}
