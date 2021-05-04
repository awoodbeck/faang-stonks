package api

import (
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/awoodbeck/faang-stonks/history"
	"github.com/awoodbeck/faang-stonks/metrics"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// newMux returns a new Gorilla mux.
func newMux(provider history.Provider, log *zap.SugaredLogger,
	instrument bool) *mux.Router {
	log = log.Named("mux")

	r := mux.NewRouter().StrictSlash(true)
	r.Use(gziphandler.GzipHandler)

	if instrument {
		r.Use(metricsMiddleware)
		log.Info("API instrumented")
	}

	s := r.Methods("GET").PathPrefix("/v1").Subrouter()
	s.HandleFunc("/stocks", stocks(provider, log))
	s.HandleFunc("/stock/{symbol:[a-zA-Z0-9]+}", stock(provider, log))

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
