package rdbms

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Shoowa/vamos/config"
	"github.com/Shoowa/vamos/secrets"
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
		secretsReader := new(secrets.SkeletonKey)
		secretsReader.Create(cfg)
		pw, pwErr := secretsReader.ReadPathAndKey(db.Secret, db.SecretKey)
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
