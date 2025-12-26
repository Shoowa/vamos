//go:build integration

package secrets_test

import (
	"testing"

	"github.com/Shoowa/vamos/config"
	. "github.com/Shoowa/vamos/secrets"
	. "github.com/Shoowa/vamos/testhelper"
)

const (
	PASSWORD     = "OpenBao123"
	REDIS_SECRET = "dev-redis-test"
	REDIS_PW     = "ReDiS4LiFe"
)

func Test_PostgresPassword(t *testing.T) {
	t.Setenv("APP_ENV", "DEV")
	t.Setenv("OPENBAO_TOKEN", "token")

	cfg := config.Read()
	db := cfg.Data.Relational[0]
	sk := new(SkeletonKey)
	sk.Create(cfg)

	pw, pwErr := sk.ReadPathAndKey(db.Secret, db.SecretKey)
	if pwErr != nil {
		t.Error(pwErr.Error())
	}

	Equals(t, PASSWORD, pw)
}

func Test_ReadValueFromOpenbao(t *testing.T) {
	t.Setenv("APP_ENV", "DEV")
	t.Setenv("OPENBAO_TOKEN", "token")

	cfg := config.Read()
	sk := new(SkeletonKey)
	sk.Create(cfg)

	val, err := sk.ReadPathAndKey(REDIS_SECRET, cfg.Cache.SecretKey)
	if err != nil {
		t.Error(err.Error())
	}

	Equals(t, REDIS_PW, val)
}

func Test_ReadIntermediateCA(t *testing.T) {
	t.Setenv("APP_ENV", "DEV")
	t.Setenv("OPENBAO_TOKEN", "token")

	cfg := config.Read()
	sk := new(SkeletonKey)
	sk.Create(cfg)

	_, err := sk.ReadIntermediateCA(cfg.HttpServer)
	Ok(t, err)
}

func Test_CreateCertPool(t *testing.T) {
	t.Setenv("APP_ENV", "DEV")
	t.Setenv("OPENBAO_TOKEN", "token")

	cfg := config.Read()
	sk := new(SkeletonKey)
	sk.Create(cfg)

	_, err := sk.CreateCertPool(cfg.HttpServer)
	Ok(t, err)
}

func Test_CreateTLSwithCA(t *testing.T) {
	t.Setenv("APP_ENV", "DEV")
	t.Setenv("OPENBAO_TOKEN", "token")

	cfg := config.Read()
	sk := new(SkeletonKey)
	sk.Create(cfg)

	tlsConfig, err := sk.ConfigureTLSwithCA(cfg)
	Ok(t, err)
	Equals(t, 1, len(tlsConfig.Certificates))
}
