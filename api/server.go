// Package api provides a HTTP(S) server that serves a REST API that serves
// financial data for a subset of stock symbols.
package api

// Server is a web server that serves up the REST API for this application.
// Currently, it allows clients to request the latest stock quotes for a given
// set of stocks.
type Server struct{}

// New returns a pointer to an API Server.
func New(options ...Option) (*Server, error) {
	s := &Server{}

	for _, option := range options {
		option(s)
	}

	return s, nil
}
