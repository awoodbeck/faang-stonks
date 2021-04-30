package iexcloud

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Option interface {
	apply(*Client)
}

type optFunc func(*Client)

func (f optFunc) apply(c *Client) {
	f(c)
}

// BatchEndpoint accepts the full URL to the IEX Cloud API to use for batchQuotes
// requests.
func BatchEndpoint(url string) Option {
	return optFunc(func(c *Client) {
		c.batchEndpoint = url
	})
}

// CallTimeout accepts a duration that the client should wait for a response
// for each call.
func CallTimeout(timeout time.Duration) Option {
	return optFunc(func(c *Client) {
		c.timeout = timeout
	})
}

// InstrumentHTTPClient replaces the default HTTP client with an instrumented
// version compatible with Prometheus.
func InstrumentHTTPClient() Option {
	return optFunc(func(c *Client) {
		inFlightGauge := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "client_in_flight_requests",
			Help: "A gauge of in-flight requests for the wrapped client.",
		})

		counter := prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "client_api_requests_total",
				Help: "A counter for requests from the wrapped client.",
			},
			[]string{"code", "method"},
		)

		// dnsLatencyVec uses custom buckets based on expected dns durations.
		// It has an instance label "event", which is set in the
		// DNSStart and DNSDonehook functions defined in the
		// InstrumentTrace struct below.
		dnsLatencyVec := prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "dns_duration_seconds",
				Help:    "Trace dns latency histogram.",
				Buckets: []float64{.005, .01, .025, .05},
			},
			[]string{"event"},
		)

		// tlsLatencyVec uses custom buckets based on expected tls durations.
		// It has an instance label "event", which is set in the
		// TLSHandshakeStart and TLSHandshakeDone hook functions defined in the
		// InstrumentTrace struct below.
		tlsLatencyVec := prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "tls_duration_seconds",
				Help:    "Trace tls latency histogram.",
				Buckets: []float64{.05, .1, .25, .5},
			},
			[]string{"event"},
		)

		// histVec has no labels, making it a zero-dimensional ObserverVec.
		histVec := prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "request_duration_seconds",
				Help:    "A histogram of request latencies.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{},
		)

		// Register all of the metrics in the standard registry.
		prometheus.MustRegister(counter, tlsLatencyVec, dnsLatencyVec, histVec,
			inFlightGauge)

		// Define functions for the available httptrace.ClientTrace hook
		// functions that we want to instrument.
		trace := &promhttp.InstrumentTrace{
			DNSStart: func(t float64) {
				dnsLatencyVec.WithLabelValues("dns_start").Observe(t)
			},
			DNSDone: func(t float64) {
				dnsLatencyVec.WithLabelValues("dns_done").Observe(t)
			},
			TLSHandshakeStart: func(t float64) {
				tlsLatencyVec.WithLabelValues("tls_handshake_start").Observe(t)
			},
			TLSHandshakeDone: func(t float64) {
				tlsLatencyVec.WithLabelValues("tls_handshake_done").Observe(t)
			},
		}

		// Wrap the default RoundTripper with middleware.
		roundTripper := promhttp.InstrumentRoundTripperInFlight(inFlightGauge,
			promhttp.InstrumentRoundTripperCounter(counter,
				promhttp.InstrumentRoundTripperTrace(trace,
					promhttp.InstrumentRoundTripperDuration(histVec,
						http.DefaultTransport,
					),
				),
			),
		)

		// Set the RoundTripper on our client.
		c.httpClient.Transport = roundTripper
	})
}
