//go:build integration

package secrets_test

import (
	"strings"
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

	pw, pwErr := sk.ReadPathAndKey(db.Secret, "password")
	if pwErr != nil {
		t.Error(pwErr.Error())
	}

	Equals(t, PASSWORD, pw)
}

func Test_SkeletonKeyOpenbao(t *testing.T) {
	t.Setenv("APP_ENV", "DEV")
	t.Setenv("OPENBAO_TOKEN", "token")

	cfg := config.Read()
	sk := new(SkeletonKey)
	sk.Create(cfg)

	addr := sk.Openbao.Address()
	tok := sk.Openbao.Token()
	Assert(t, strings.Contains(addr, "localhost"), "Lacks localhost in host address.")
	Assert(t, strings.Contains(tok, "token"), "Lacks token from environment.")
}

func Test_ReadValueFromOpenbao(t *testing.T) {
	t.Setenv("APP_ENV", "DEV")
	t.Setenv("OPENBAO_TOKEN", "token")

	cfg := config.Read()
	sk := new(SkeletonKey)
	sk.Create(cfg)

	val, err := sk.ReadPathAndKey(REDIS_SECRET, "password")
	if err != nil {
		t.Error(err.Error())
	}

	Equals(t, REDIS_PW, val)
}
