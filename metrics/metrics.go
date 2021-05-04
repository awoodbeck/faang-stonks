// Package metrics provides instrumentation for the rest of the application,
// and the requisite HTTP server to serve the Prometheus endpoint separate
// from the REST API endpoint.
package metrics

import "github.com/prometheus/client_golang/prometheus"

func init() {
	prometheus.MustRegister(
		ClientAPIRequests,
		ClientDNSDuration,
		ClientInFlightRequests,
		ClientRequestDuration,
		ClientTLSDuration,
		ServerAPIRequests,
		ServerInFlightRequests,
		ServerRequestDuration,
		ServerResponseBytes,
	)
}

// ClientAPIRequests keeps count of the number of API requests made by the
// HTTP client.
var ClientAPIRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "client_api_requests_total",
		Help: "A counter for requests from the wrapped client.",
	}, []string{},
)

// ClientDNSDuration tracks DNS latency.
var ClientDNSDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "client_dns_duration_seconds",
		Help:    "Trace DNS latency histogram.",
		Buckets: prometheus.DefBuckets,
	}, []string{},
)

// ClientInFlightRequests tracks the number of in-flight client requests.
var ClientInFlightRequests = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "client_in_flight_requests",
		Help: "A gauge of in-flight requests for the wrapped client.",
	},
)

// ClientRequestDuration tracks client request durations.
var ClientRequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "client_request_duration_seconds",
		Help:    "A histogram of request latencies.",
		Buckets: prometheus.DefBuckets,
	}, []string{},
)

// ClientTLSDuration tracks TLS latency.
var ClientTLSDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "client_tls_duration_seconds",
		Help:    "Trace TLS latency histogram.",
		Buckets: prometheus.DefBuckets,
	}, []string{},
)

// ServerAPIRequests counts the number of API requests handled by the server.
var ServerAPIRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "api_requests_total",
		Help: "A counter for requests to the wrapped handler.",
	},
	[]string{},
)

// ServerInFlightRequests tracks the number of in-flight requests into the
// API server.
var ServerInFlightRequests = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "server_in_flight_requests",
		Help: "A gauge of requests currently being served by the wrapped handler.",
	},
)

// ServerRequestDuration tracks the request latency between the client and
// API server.
var ServerRequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "request_duration_seconds",
		Help:    "A histogram of latencies for requests.",
		Buckets: prometheus.DefBuckets,
	},
	[]string{},
)

// ServerResponseBytes tracks the response payload sizes.
var ServerResponseBytes = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "response_size_bytes",
		Help:    "A histogram of response sizes for requests.",
		Buckets: prometheus.DefBuckets,
	},
	[]string{},
)
