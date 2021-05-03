// Package sqlite provides history.Archiver and history.Provider
// implementations that use SQLite for storage.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/awoodbeck/faang-stonks/finance"
	"github.com/awoodbeck/faang-stonks/history"
	_ "github.com/mattn/go-sqlite3"
)

const (
	defaultDBFile = "stonks.sqlite"

	// TODO: I can make an argument for and against normalizing the symbols
	// column. I'll keep it as-is for the purposes of this demo.
	createQuotesTable = `
CREATE TABLE "quotes"
(
	id integer not null
		constraint quotes_pk
			primary key autoincrement,
	symbol text not null,
	price real not null,
	datetime timestamp not null
)`

	insertQuote = `
INSERT INTO quotes (symbol, price, datetime)
  VALUES (?, ?, ?)`

	selectQuotes = `
SELECT symbol, price, datetime
  FROM quotes
  WHERE symbol = ?
  ORDER BY id DESC
  LIMIT ?`
)

var (
	defaultSymbols = []string{"fb", "amzn", "aapl", "nflx", "goog"}

	_ history.Archiver = (*Client)(nil)
	_ history.Provider = (*Client)(nil)
)

// Client implements the history.Archiver and history.Provider interfaces,
// knowing how to store and retrieve stock quotes, respectively.
type Client struct {
	db      *sql.DB
	file    string
	symbols map[string]struct{}
}

// initialize the database file.
func (c *Client) initialize() error {
	// TODO: Add proper support for data migrations. For demo purposes, we
	// start with a fresh database upon each invocation of this application.
	err := os.Remove(c.file)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing %q: %w", c.file, err)
	}

	c.db, err = sql.Open("sqlite3", c.file)
	if err != nil {
		return fmt.Errorf("open %q: %w", c.file, err)
	}

	_, err = c.db.Exec(createQuotesTable)
	if err != nil {
		return fmt.Errorf("creating quotes table: %w", err)
	}

	return nil
}

// Close the database connection.
func (c Client) Close() error {
	if c.db == nil {
		return nil
	}

	return c.db.Close()
}

// GetQuotes accepts a stock symbol and the latest quotes for the stock to
// return.
func (c Client) GetQuotes(ctx context.Context, symbol string, last int) (
	[]finance.Quote, error) {
	stmt, err := c.db.Prepare(selectQuotes)
	if err != nil {
		return nil, fmt.Errorf("selecting quotes: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	if last < 1 {
		last = 1
	}

	rows, err := stmt.Query(strings.ToLower(symbol), last)
	if err != nil {
		return nil, fmt.Errorf("select query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	quotes := make([]finance.Quote, 0, last)

	for rows.Next() {
		var (
			q finance.Quote
			t time.Time
		)
		err = rows.Scan(&q.Symbol, &q.Price, &t)
		if err != nil {
			return nil, fmt.Errorf("row scan: %w", err)
		}
		q.Time = t.Local()

		quotes = append(quotes, q)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return quotes, nil
}

// SetQuotes accepts a slice of finance.Quote objects and archives them to
// SQLite.
func (c Client) SetQuotes(ctx context.Context, quotes []finance.Quote) error {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, insertQuote)
	if err != nil {
		return fmt.Errorf("preparing insert: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, q := range quotes {
		_, err = stmt.Exec(strings.ToLower(q.Symbol), q.Price, q.Time.UTC())
		if err != nil {
			return fmt.Errorf("inserting %v: %w", q, err)
		}
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

// New returns a pointer to a new Client object after applying optional settings.
func New(options ...Option) (*Client, error) {
	c := &Client{
		file:    defaultDBFile,
		symbols: make(map[string]struct{}),
	}

	for _, symbol := range defaultSymbols {
		c.symbols[symbol] = struct{}{}
	}

	c.db.SetMaxIdleConns(2)
	c.db.SetConnMaxLifetime(-1)

	for _, option := range options {
		option(c)
	}

	if err := c.initialize(); err != nil {
		return nil, err
	}

	return c, nil
}
