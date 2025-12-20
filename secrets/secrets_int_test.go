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
	PASSWORD = "OpenBao123"
)

func Test_Connection(t *testing.T) {
	t.Setenv("APP_ENV", "DEV")
	t.Setenv("OPENBAO_TOKEN", "token")

	cfg := config.Read()
	db := cfg.Data.Relational[0]

	pw, pwErr := BuildAndRead(cfg, db.Secret)
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
