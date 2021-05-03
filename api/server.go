// Package api provides a HTTP(S) server that serves a REST API that serves
// financial data for a subset of stock symbols.
package api

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
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

func (s *Server) ListenAndServe() error {
	go func() {
		<-s.ctx.Done()
		_ = s.srv.Close()
	}()

	return s.srv.ListenAndServe()
}

func (s *Server) ListenAndServeTLS(cert, pkey string) error {
	go func() {
		<-s.ctx.Done()
		_ = s.srv.Close()
	}()

	return s.srv.ListenAndServeTLS(cert, pkey)
}

// New returns a pointer to an API Server.
func New(ctx context.Context, log *zap.SugaredLogger, options ...Option) (
	*Server, error) {
	s := &Server{
		ctx:               ctx,
		log:               log,
		listenAddr:        ":18081",
		idleTimeout:       time.Minute,
		readHeaderTimeout: 30 * time.Second,
		instrumentation:   true,
	}

	for _, option := range options {
		option(s)
	}

	s.srv = &http.Server{
		IdleTimeout:       s.idleTimeout,
		ReadHeaderTimeout: s.readHeaderTimeout,
		Handler:           NewMux(s),
	}

	return s, nil
}
