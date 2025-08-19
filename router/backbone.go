package router

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	StatusClientClosed = 499
)

type errHandler func(http.ResponseWriter, *http.Request) error

type Option func(*Backbone)

// Backbone holds data dependencies. The field FirstDB can hold a sqlC generated
// Queries struct, but we can't define that type.
type Backbone struct {
	Logger       *slog.Logger
	Health       *Health
	DbHandle     *pgxpool.Pool
	HeapSnapshot *bytes.Buffer
}

// The Options pattern is used to configure the struct, because the struct
// itself will be defined differently depending on a build tag.
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

func WithLogger(l *slog.Logger) Option {
	return func(b *Backbone) {
		b.Logger = l
	}
}

func WithDbHandle(dbHandle *pgxpool.Pool) Option {
	return func(b *Backbone) {
		b.DbHandle = dbHandle
	}
}

func (b *Backbone) eHand(f errHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		err := f(w, req)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				b.Logger.Error("HTTP", "status", StatusClientClosed)
			case errors.Is(err, context.DeadlineExceeded):
				b.Logger.Error("HTTP", "status", http.StatusRequestTimeout)
				http.Error(w, "timeout", http.StatusRequestTimeout)
			case errors.Is(err, sql.ErrNoRows):
				w.WriteHeader(http.StatusNoContent)
			default:
				b.Logger.Error("HTTP", "err", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	}
}
