//go:build integration

package secrets_test

import (
	"testing"

	. "github.com/Shoowa/vamos/secrets"
	. "github.com/Shoowa/vamos/testhelper"
)

const (
	PASSWORD     = "OpenBao123"
	REDIS_SECRET = "dev-redis-test"
	REDIS_PW     = "ReDiS4LiFe"
)

func Test_PostgresPassword(t *testing.T) {
	db := cfg.Data.Relational[0]
	sk := new(SkeletonKey)
	sk.Create(cfg)

	pw, pwErr := sk.ReadPathAndKey(db.Secret, db.SecretKey)

	Ok(t, pwErr)
	Equal(t, PASSWORD, pw)
}

func Test_ReadValueFromOpenbao(t *testing.T) {
	sk := new(SkeletonKey)
	sk.Create(cfg)

	val, err := sk.ReadPathAndKey(REDIS_SECRET, cfg.Cache.SecretKey)

	Ok(t, err)
	Equal(t, REDIS_PW, val)
}

func Test_ReadIntermediateCA(t *testing.T) {
	sk := new(SkeletonKey)
	sk.Create(cfg)

	_, err := sk.ReadIntermediateCA(cfg.HttpServer)
	Ok(t, err)
}

func Test_CreateCertPool(t *testing.T) {
	sk := new(SkeletonKey)
	sk.Create(cfg)

	_, err := sk.CreateCertPool(cfg.HttpServer)
	Ok(t, err)
}

func Test_CreateTLSwithCA(t *testing.T) {
	sk := new(SkeletonKey)
	sk.Create(cfg)

	tlsConfig, err := sk.ConfigureTLSwithCA(cfg.HttpServer)
	Ok(t, err)
	Equal(t, 1, len(tlsConfig.Certificates))
}
