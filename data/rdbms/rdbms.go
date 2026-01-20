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

// WhichDB reads from a list of databases in the Config struct. A developer must
// select an index in that array.
func WhichDB(cfg *config.Config, dbPosition int) config.Rdb {
	return cfg.Data.Relational[dbPosition]
}

func sslMode(flag bool) string {
	if flag == true {
		return "verify-full"
	} else {
		return "disable"
	}
}

// Credentials conveniently assembles a string for the Postgres client.
func Credentials(db config.Rdb) (string, error) {
	credString := fmt.Sprintf(
		"user=%v host=%v database=%v sslmode=%v",
		db.User, db.Host, db.Database, sslMode(db.Sslmode),
	)
	return credString, nil
}

// configure chooses a database from an array in the Config file, and then adds
// the capability to read a password from secret storage any time, and adds TLS.
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

	if db.Sslmode == true {
		// temporary SecretsReader creation that will die out of scope.
		secReader := new(secrets.SkeletonKey)
		secReader.Create(cfg)

		// Read certificate, key, CA from Secrets storage.
		tlsInfo, tlsErr := secReader.ConfigureTLSwithCA(cfg.HttpServer)
		if tlsErr != nil {
			panic(tlsErr.Error())
		}

		// Add expected hostname of Postgres server.
		tlsInfo.ServerName = db.Host

		// Configure connection pool with TLS.
		pgxConfig.ConnConfig.TLSConfig = tlsInfo
	}

	return pgxConfig, nil
}

// ConnectDB configures and creates a Postgres connection pool.
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
