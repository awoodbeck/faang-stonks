package api

import (
	"net/http"

	"github.com/awoodbeck/faang-stonks/metrics"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewMux returns a new Gorilla mux
func NewMux(s *Server) *mux.Router {
	r := mux.NewRouter()

	if s.instrumentation {
		r.Use(metricsMiddleware)
	}

	return r
}

func metricsMiddleware(next http.Handler) http.Handler {
	return promhttp.InstrumentHandlerInFlight(metrics.ServerInFlightRequests,
		promhttp.InstrumentHandlerDuration(metrics.ServerRequestDuration,
			promhttp.InstrumentHandlerCounter(metrics.ServerAPIRequests,
				promhttp.InstrumentHandlerResponseSize(metrics.ServerResponseBytes,
					next,
				),
			),
		),
	)
}
