// Package api provides a HTTP(S) server that serves a REST API that serves
// financial data for a subset of stock symbols.
//
// In the absence of additional middleware to handle details such as
// compression, this server assumes those transport details are handled by
// a reverse proxy.
package api

import (
	"context"
	"net/http"
	"time"

	"github.com/awoodbeck/faang-stonks/history"
	"go.uber.org/zap"
)

const (
	// DefaultIdleTimeout represents the duration the server will allow clients
	// to remain idle.
	DefaultIdleTimeout = time.Minute

	// DefaultListenAddress is the default host:port on which the API server
	// will listen for client connections.
	DefaultListenAddress = ":18081"

	// DefaultReadHeaderTimeout is the default duration during which the API
	// server will wait for the client to send request headers.
	DefaultReadHeaderTimeout = 30 * time.Second
)

// Server is a web server that serves up the REST API for this application.
// Currently, it allows clients to request the latest stock quotes for a given
// set of stocks.
type Server struct {
	ctx               context.Context
	srv               *http.Server
	log               *zap.SugaredLogger
	listenAddr        string
	idleTimeout       time.Duration
	readHeaderTimeout time.Duration
	instrumentation   bool
}

// ListenAndServe binds the server to its address and serves incoming requests.
func (s *Server) ListenAndServe() error {
	go func() {
		<-s.ctx.Done()
		s.log.Info("shutting down ...")
		sCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = s.srv.Shutdown(sCtx)
	}()

	return s.srv.ListenAndServe()
}

// ListenAndServeTLS binds the server to its address and serves incoming
// requests over TLS.
//
// Note: This is best handled by a reverse proxy like Nginx or Caddy,
// particularly one that speaks the ACME protocol and can make TLS integration
// virtually effortless.
func (s *Server) ListenAndServeTLS(cert, pkey string) error {
	go func() {
		<-s.ctx.Done()
		s.log.Info("shutting down ...")
		sCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = s.srv.Shutdown(sCtx)
	}()

	return s.srv.ListenAndServeTLS(cert, pkey)
}

// New returns a pointer to an API Server.
//
// Defaults:
//     IdleTimeout       = time.Minute
//     ListenAddress     = ":18081"
//     ReadHeaderTimeout = 30 * time.Second
func New(ctx context.Context, p history.Provider, log *zap.SugaredLogger,
	options ...Option) (
	*Server, error) {
	s := &Server{
		ctx:               ctx,
		log:               log.Named("api"),
		listenAddr:        DefaultListenAddress,
		idleTimeout:       DefaultIdleTimeout,
		readHeaderTimeout: DefaultReadHeaderTimeout,
		instrumentation:   true,
	}

	for _, option := range options {
		if option != nil {
			option(s)
		}
	}

	s.srv = &http.Server{
		IdleTimeout:       s.idleTimeout,
		ReadHeaderTimeout: s.readHeaderTimeout,
		Handler:           newMux(p, log, s.instrumentation),
	}

	return s, nil
}
