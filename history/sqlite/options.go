package sqlite

import "time"

type Option func(*Client)

// ConnMaxLifetime sets the maximum lifetime of each connection to the given
// duration.
func ConnMaxLifetime(d time.Duration) Option {
	return func(c *Client) {
		c.db.SetConnMaxLifetime(d)
	}
}

// DatabaseFile specifies the database file name to use.
func DatabaseFile(f string) Option {
	return func(c *Client) {
		c.file = f
	}
}

// MaxIdleConnections sets the maximum number of connections allowed to remain
// idle.
func MaxIdleConnections(i int) Option {
	return func(c *Client) {
		c.db.SetMaxIdleConns(i)
	}
}

// Symbols configures the Archiver to track specific stock symbols.
func Symbols(symbols []string) Option {
	return func(c *Client) {
		c.symbols = make(map[string]struct{})

		for _, symbol := range symbols {
			c.symbols[symbol] = struct{}{}
		}
	}
}
