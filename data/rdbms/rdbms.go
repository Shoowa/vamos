package rdbms

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"vamos/config"
	"vamos/secrets"
	"vamos/sqlc/data/first"
)

const (
	TIMEOUT_PING = time.Second * 1
)

func WhichDB(cfg *config.Config, dbPosition int) config.Rdb {
	return cfg.Data.Relational[dbPosition]
}

func Credentials(db config.Rdb) (string, error) {
	credString := fmt.Sprintf(
		"user=%v host=%v database=%v sslmode=%v",
		db.User, db.Host, db.Database, db.Sslmode,
	)
	return credString, nil
}

func configure(cfg *config.Config, dbPosition int) (*pgxpool.Config, error) {
	db := WhichDB(cfg, dbPosition)

	credString, credErr := Credentials(db)
	if credErr != nil {
		return nil, credErr
	}

	pgxConfig, configErr := pgxpool.ParseConfig(credString)
	if configErr != nil {
		return nil, configErr
	}

	pgxConfig.BeforeConnect = func(ctx context.Context, cc *pgx.ConnConfig) error {
		pw, pwErr := secrets.BuildAndRead(cfg, db.Secret)
		if pwErr != nil {
			return pwErr
		}

		cc.Password = pw
		return nil
	}

	return pgxConfig, nil
}

func ConnectDB(cfg *config.Config, dbPosition int) (*pgxpool.Pool, error) {
	dbConfig, dbConfigErr := configure(cfg, dbPosition)
	if dbConfigErr != nil {
		return nil, dbConfigErr
	}

	ctxTimer, cancel := context.WithTimeout(context.Background(), TIMEOUT_PING)
	defer cancel()

	dbpool, connErr := pgxpool.NewWithConfig(ctxTimer, dbConfig)

	if connErr != nil {
		return nil, connErr
	}

	pingErr := dbpool.Ping(ctxTimer)
	if pingErr != nil {
		return nil, pingErr
	}

	return dbpool, nil
}

func FirstDB_AdoptQueries(dbpool *pgxpool.Pool) *first.Queries {
	return first.New(dbpool)
}
