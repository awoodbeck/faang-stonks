// Package influxdb provides history.Archiver and history.Provider
// implementations that use the time series database, InfluxDB.
//
// Note: This stub is included as an example of how I could add InfluxDB
// support if required in the future. Abstracting this away from the rest of
// the code allows me to transparently swap backend implementations (e.g.,
// SQLite for InfluxDB) as requirements and scaling needs change. For the
// purposes of this demo, I opted to keep things simple and use SQLite.
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

func (c Client) Close() error {
	panic("implement me")
}

func (c Client) GetQuotes(ctx context.Context, symbol string, last int) (
	[]finance.Quote, error) {
	panic("implement me")
}

func (c Client) SetQuotes(ctx context.Context, quotes []finance.Quote) error {
	panic("implement me")
}
