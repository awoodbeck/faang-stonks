package api

import "time"

type Option func(*Server)

// DisableInstrumentation turns off server instrumentation.
func DisableInstrumentation() Option {
	return func(s *Server) {
		s.instrumentation = false
	}
}

// IdleTimeout sets the server's IdleTimeout value to the given duration.
func IdleTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.idleTimeout = d
	}
}

// ListenAddress specifies the server's listen address.
func ListenAddress(addr string) Option {
	return func(s *Server) {
		s.listenAddr = addr
	}
}

// ReadHeaderTimeout sets the server's ReadHeaderTimeout value to the given
// duration.
func ReadHeaderTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.readHeaderTimeout = d
	}
}
