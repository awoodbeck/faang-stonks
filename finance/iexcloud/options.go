package iexcloud

import (
	"net/http"
	"time"

	"github.com/awoodbeck/faang-stonks/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Option func(*Client)

// BatchEndpoint accepts the full URL to the IEX Cloud API to use for batch
// requests.
func BatchEndpoint(url string) Option {
	return func(c *Client) {
		c.batchEndpoint = url
	}
}

// CallTimeout accepts a duration that the client should wait for a response
// for each call.
func CallTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// InstrumentHTTPClient replaces the default HTTP client with an instrumented
// version compatible with Prometheus.
func InstrumentHTTPClient() Option {
	return func(c *Client) {
		// Define functions for the available httptrace.ClientTrace hook
		// functions that we want to instrument.
		trace := &promhttp.InstrumentTrace{
			DNSStart: func(t float64) {
				metrics.ClientDNSDuration.WithLabelValues("dns_start").Observe(t)
			},
			DNSDone: func(t float64) {
				metrics.ClientDNSDuration.WithLabelValues("dns_done").Observe(t)
			},
			TLSHandshakeStart: func(t float64) {
				metrics.ClientTLSDuration.WithLabelValues("tls_handshake_start").Observe(t)
			},
			TLSHandshakeDone: func(t float64) {
				metrics.ClientTLSDuration.WithLabelValues("tls_handshake_done").Observe(t)
			},
		}

		// Wrap the default RoundTripper with middleware.
		roundTripper := promhttp.InstrumentRoundTripperInFlight(metrics.ClientInFlightRequests,
			promhttp.InstrumentRoundTripperCounter(metrics.ClientAPIRequests,
				promhttp.InstrumentRoundTripperTrace(trace,
					promhttp.InstrumentRoundTripperDuration(metrics.ClientRequestDuration,
						http.DefaultTransport,
					),
				),
			),
		)

		// Set the RoundTripper on our client.
		c.httpClient.Transport = roundTripper
	}
}
