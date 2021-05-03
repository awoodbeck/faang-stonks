package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// New returns a pointer to a http.Server instance configured to serve
// Prometheus metrics on TCP port 2112, by default.
func New(options ...Option) *http.Server {
	s := &http.Server{
		Addr:              ":2112",
		Handler:           promhttp.Handler(),
		IdleTimeout:       time.Minute,
		ReadHeaderTimeout: 30 * time.Second,
	}

	for _, option := range options {
		option(s)
	}

	return s
}
