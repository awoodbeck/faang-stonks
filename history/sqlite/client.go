// Package sqlite provides history.Archiver and history.Provider
// implementations that use SQLite for storage.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/awoodbeck/faang-stonks/finance"
	"github.com/awoodbeck/faang-stonks/history"
	_ "github.com/mattn/go-sqlite3"
)

const (
	defaultDBFile      = "stonks.sqlite"
	createSymbolsTable = `
CREATE TABLE "symbols"
(
	id integer not null
		constraint stonks_pk
			primary key autoincrement,
	symbol text not null
)`
	createQuotesTable = `
CREATE TABLE quotes
(
	id integer not null
		constraint quotes_pk
			primary key autoincrement,
	symbol integer not null
		constraint quotes_stonks_id_fk
			references "symbols",
	price real not null,
	timestamp numeric not null
)`
	insertSymbols = `INSERT INTO symbols (symbol) VALUES (?);`
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
	symbols []string
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

	_, err = c.db.Exec(createSymbolsTable)
	if err != nil {
		return fmt.Errorf("creating symbols table: %w", err)
	}

	_, err = c.db.Exec(createQuotesTable)
	if err != nil {
		return fmt.Errorf("creating quotes table: %w", err)
	}

	tx, err := c.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}

	stmt, err := tx.Prepare(insertSymbols)
	if err != nil {
		return fmt.Errorf("preparing insert: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, symbol := range c.symbols {
		_, err = stmt.Exec(symbol)
		if err != nil {
			return fmt.Errorf("inserting %q: %w", symbol, err)
		}
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("committing transaction: %w", err)
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
	panic("implement me")
}

// SetQuotes accepts a slice of finance.Quote objects and archives them to
// SQLite.
func (c Client) SetQuotes(ctx context.Context, quotes []finance.Quote) error {
	panic("implement me")
}

// New returns a pointer to a new Client object after applying optional settings.
func New(options ...Option) (*Client, error) {
	c := &Client{
		file:    defaultDBFile,
		symbols: defaultSymbols,
	}

	for _, option := range options {
		option(c)
	}

	if err := c.initialize(); err != nil {
		return nil, err
	}

	return c, nil
}
