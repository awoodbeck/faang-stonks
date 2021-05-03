package metrics

import (
	"net/http"
	"time"
)

type Option func(*http.Server)

// IdleTimeout sets the server's IdleTimeout value to the given duration.
func IdleTimeout(d time.Duration) Option {
	return func(s *http.Server) {
		s.IdleTimeout = d
	}
}

// ListenAddress specifies the server's listen address.
func ListenAddress(addr string) Option {
	return func(s *http.Server) {
		s.Addr = addr
	}
}

// ReadHeaderTimeout sets the server's ReadHeaderTimeout value to the given
// duration.
func ReadHeaderTimeout(d time.Duration) Option {
	return func(s *http.Server) {
		s.ReadHeaderTimeout = d
	}
}
