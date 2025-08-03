package server

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"vamos/internal/data/rdbms"
	"vamos/sqlc/data/first"
)

type Option func(*Backbone)

// Backbone holds data dependencies.
type Backbone struct {
	Logger  *slog.Logger
	FirstDB *first.Queries
	Health  *Health
	DbHandle *pgxpool.Pool
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
	b.Health = health
	return b
}

func WithLogger(l *slog.Logger) Option {
	return func(b *Backbone) {
		b.Logger = l
	}
}

func WithQueryHandleForFirstDB(dbHandle *pgxpool.Pool) Option {
	return func(b *Backbone) {
		q := rdbms.FirstDB_AdoptQueries(dbHandle)
		b.FirstDB = q
	}
}

func WithDbHandle(dbHandle *pgxpool.Pool) Option {
	return func(b *Backbone) {
		b.DbHandle = dbHandle
	}
}
