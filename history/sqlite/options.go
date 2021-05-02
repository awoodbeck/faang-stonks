package sqlite

type Option func(*Client)

// DatabaseFile specifies the database file name to use.
func DatabaseFile(f string) Option {
	return func(c *Client) {
		c.file = f
	}
}

// Symbols configures the Archiver to track specific stock symbols.
func Symbols(symbols []string) Option {
	s := make([]string, len(symbols))
	copy(s, symbols)
	return func(c *Client) {
		c.symbols = s
	}
}
