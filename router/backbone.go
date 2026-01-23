package router

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	redis "github.com/redis/go-redis/v9"
)

const (
	StatusClientClosed = 499
)

// Gatherer is an ugly, contrived interface that this library's configured
// router expects. So it is adopted by the native Backbone struct. It can also
// be adopted by a struct that wraps around Backbone in a downstream executable.
// Any Gatherer can be passed to the function that creates a new router. This
// enabled creating a custom test server in the library that downstream
// developers can easily use. Using an interface allows me to pass a
// dependency-wrapper around both a library and a downstream executable, and for
// production and testing.
type Gatherer interface {
	GetLogger() *slog.Logger
	GetBackbone() *Backbone
	AddBackbone(*Backbone)
	GetEndpoints() []Endpoint
}

// Option allows us to selectively add items to the Backbone struct.
type Option func(*Backbone)

// Backbone holds dependencies that can eventually be accessed by a
// http.Handler.
type Backbone struct {
	Logger       *slog.Logger
	Health       *Health
	DbHandle     *pgxpool.Pool
	Cache        *redis.Client
	HeapSnapshot *bytes.Buffer
}

// NewBackbone employs the Options pattern to selectively configure the Backbone
// struct. It also adds a Health record and initializes it. It also adds a
// buffer for receiving runtime profile data.
func NewBackbone(options ...Option) *Backbone {
	b := new(Backbone)
	for _, opt := range options {
		opt(b)
	}
	health := new(Health)
	health.Rdbms = false
	health.Heap = true
	health.Routines = true
	b.Health = health
	buf := new(bytes.Buffer)
	b.HeapSnapshot = buf
	return b
}

// WithLogger selectively adds a structured logger to the Backbone struct.
func WithLogger(l *slog.Logger) Option {
	return func(b *Backbone) {
		b.Logger = l
	}
}

// WithDbHandle selectively adds a Postgres connection pool to the Backbone
// struct.
func WithDbHandle(dbHandle *pgxpool.Pool) Option {
	return func(b *Backbone) {
		b.DbHandle = dbHandle
	}
}

// WithCache selectively adds a Redis client to the Backbone struct.
func WithCache(client *redis.Client) Option {
	return func(b *Backbone) {
		b.Cache = client
	}
}

// ServerError logs an error, then produces a HTTP response appropriate for
// common errors.
func (b *Backbone) ServerError(w http.ResponseWriter, r *http.Request, err error) {
	method := r.Method
	path := r.URL.Path

	switch {
	case errors.Is(err, context.Canceled):
		b.Logger.Warn("HTTP", "status", StatusClientClosed, "method", method, "path", path)
	case errors.Is(err, context.DeadlineExceeded):
		b.Logger.Error("HTTP", "status", http.StatusGatewayTimeout, "method", method, "path", path)
		http.Error(w, "timeout", http.StatusGatewayTimeout)
	case errors.Is(err, sql.ErrNoRows):
		w.WriteHeader(http.StatusNoContent)
	default:
		b.Logger.Error("HTTP", "err", err.Error(), "method", method, "path", path)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetLogger is a convoluted method on the Backbone struct that fulfills the
// Gatherer interface.
func (b *Backbone) GetLogger() *slog.Logger {
	return b.Logger
}

// GetBackbone is an awkward, convoluted method on the Backbone struct to summon
// itself. This awkwardly satisfies the Gatherer interface, and using an
// interface allows me to pass a dependency-wrapper around both a library and a
// downstream executable, and for production and for testing.
func (b *Backbone) GetBackbone() *Backbone {
	return b
}

// AddBackbone is an awkward method to satisfy the Gatherer interface.
func (b *Backbone) AddBackbone(*Backbone) {}

// GetEndpoints is a convenient method to add a list of HTTP methods and
// http.Handlers to a router.
func (b *Backbone) GetEndpoints() []Endpoint {
	return nil
}
