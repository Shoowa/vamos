// Package SERVER provides a sanely configured webserver.
package server

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Shoowa/vamos/config"
	"github.com/Shoowa/vamos/router"
)

const GRACE_PERIOD = time.Second * 15

// addGlobalRateLimiter coditionally applies a rate limiting middleware to a router.
func addGlobalRateLimiter(cfg *config.RateLimiter, rtr http.Handler) http.Handler {
	if cfg.Active == true {
		limiter := router.CreateRateLimiter(cfg)
		return router.Limit(limiter, rtr)
	}
	return rtr
}

// NewServer creates a custom http.Server struct. And transfers dependencies in
// Backbone to the routing layer, so that HTTP Handlers can access a logger, a
// database, & a cache.
//
// A cancel-able parent context is supplied to the http.Server in BaseContext,
// and will be shared across all inbound requests.  The associated cancelFunc
// will notify the HTTP Handlers to terminate active connections when the server
// is ordered to halt.
//
// The server receives a x509 certificate and adopts TLS 1.3
func NewServer(cfg *config.Config, router http.Handler, cert *tls.Certificate, slogger *slog.Logger) *http.Server {
	base, stop := context.WithCancel(context.Background())
	s := &http.Server{
		Addr:         ":" + cfg.HttpServer.Port,
		Handler:      addGlobalRateLimiter(cfg.HttpServer.GlobalRateLimiter, router),
		ErrorLog:     slog.NewLogLogger(slogger.Handler(), slog.LevelError),
		ReadTimeout:  time.Second * time.Duration(cfg.HttpServer.TimeoutRead),
		WriteTimeout: time.Second * time.Duration(cfg.HttpServer.TimeoutWrite),
		IdleTimeout:  time.Second * time.Duration(cfg.HttpServer.TimeoutIdle),
		BaseContext:  func(lstnr net.Listener) context.Context { return base },
		TLSConfig: &tls.Config{
			MinVersion:   tls.VersionTLS13,
			Certificates: []tls.Certificate{*cert},
		},
	}
	s.RegisterOnShutdown(stop)
	return s
}

// gracefulIgnition launches a webserver.
func gracefulIgnition(s *http.Server) {
	err := s.ListenAndServeTLS("", "")
	if !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}

// gracefulShutdown stops accepting new connections and waits for working
// connections to become idle before terminating them.
func gracefulShutdown(s *http.Server) error {
	quitCtx, quit := context.WithTimeout(context.Background(), GRACE_PERIOD)
	defer quit()

	err := s.Shutdown(quitCtx)
	if err != nil {
		return err
	}
	return nil
}

// catchSigTerm creates a buffered message queue awaiting an OS signal. The Main
// routine will block while the channel awaits the signal. After receiving a
// signal, the Main routine will shutdown the server.
func catchSigTerm() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}

// Start launches the webserver in a go routine, then awaits a signal to
// gracefully terminate the webserver. When the webserver fails to gracefully
// stop, then it will be killed.
func Start(l *slog.Logger, s *http.Server) {
	go gracefulIgnition(s)
	l.Info("HTTP Server activated")
	catchSigTerm()
	l.Info("Begin decommissioning HTTP server.")
	shutErr := gracefulShutdown(s)
	if shutErr != nil {
		l.Error("HTTP Server shutdown error", "ERR:", shutErr.Error())
		killErr := s.Close()
		if killErr != nil {
			l.Error("HTTP Server kill error", "ERR:", killErr.Error())
		}
	}
	l.Info("HTTP Server halted")
}
