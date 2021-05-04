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

	// partition by symbol and select the top N rows from each partition
	// when batching
	selectQuotesBatch = `
WITH summary AS (
  SELECT q.symbol, q.price, q.datetime, ROW_NUMBER()
    OVER(PARTITION BY q.symbol
    ORDER BY q.id DESC) AS rank
  FROM quotes q
)
SELECT s.*
FROM summary s
WHERE symbol IN (XXX)
  AND s.rank <= ?`
)

var (
	_ history.Archiver = (*Client)(nil)
	_ history.Provider = (*Client)(nil)
)

// Client implements the history.Archiver and history.Provider interfaces,
// knowing how to store and retrieve stock quotes, respectively.
type Client struct {
	db              *sql.DB
	file            string
	maxIdleConns    int
	maxConnLifetime time.Duration
	symbols         map[string]struct{}
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

	if len(quotes) == 0 {
		return nil, history.ErrNotFound
	}

	return quotes, nil
}

// GetQuotesBatch accepts a slice of symbols and an integer indicating the last
// N quotes per symbol to return to the caller.
func (c *Client) GetQuotesBatch(ctx context.Context, symbols []string,
	last int) (finance.QuoteBatch, error) {

	// We need to build up the query string to ensure it includes the correct
	// number of place holders per symbol in the IN clause. It's not pretty,
	// but it's more attractive than little Bobby Tables (xkcd #327 for the
	// reference).
	q := fmt.Sprintf("?%s", strings.Repeat(", ?", len(symbols)-1))
	stmt, err := c.db.Prepare(strings.Replace(selectQuotesBatch, "XXX", q, 1))
	if err != nil {
		return nil, fmt.Errorf("selecting quotes batch: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	if last < 1 {
		last = 1
	}

	lcSymbols := make([]interface{}, 0, len(symbols)+1)
	for _, symbol := range symbols {
		lcSymbols = append(lcSymbols, strings.ToLower(symbol))
	}
	lcSymbols = append(lcSymbols, last)

	rows, err := stmt.Query(lcSymbols...)
	if err != nil {
		return nil, fmt.Errorf("select query batch: %w", err)
	}
	defer func() { _ = rows.Close() }()

	batch := make(finance.QuoteBatch)

	for rows.Next() {
		var (
			q finance.Quote
			t time.Time
			r int
		)
		err = rows.Scan(&q.Symbol, &q.Price, &t, &r)
		if err != nil {
			return nil, fmt.Errorf("row scan: %w", err)
		}
		q.Time = t.Local()

		batch[q.Symbol] = append(batch[q.Symbol], q)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	if len(batch) == 0 {
		return nil, history.ErrNotFound
	}

	return batch, nil
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
		file:            defaultDBFile,
		maxConnLifetime: -1,
		maxIdleConns:    2,
		symbols:         make(map[string]struct{}),
	}

	for _, symbol := range finance.DefaultSymbols {
		c.symbols[symbol] = struct{}{}
	}

	for _, option := range options {
		option(c)
	}

	if err := c.initialize(); err != nil {
		return nil, err
	}

	c.db.SetConnMaxLifetime(c.maxConnLifetime)
	c.db.SetMaxIdleConns(c.maxIdleConns)

	return c, nil
}
