//go:build integration

package secrets_test

import (
	"testing"

	"vamos/config"
	. "vamos/secrets"
	. "vamos/testhelper"
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
