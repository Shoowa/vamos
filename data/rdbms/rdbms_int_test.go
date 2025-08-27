//go:build integration

package rdbms_test

import (
	"os"
	"testing"

	"github.com/Shoowa/vamos/config"
	. "github.com/Shoowa/vamos/data/rdbms"
	. "github.com/Shoowa/vamos/testhelper"
)

func TestMain(m *testing.M) {
	os.Setenv("APP_ENV", "DEV")
	Change_to_project_root()
	os.Unsetenv("APP_ENV")

	code := m.Run()
	os.Exit(code)
}

// First, test opening a connection to a database.
func Test_ConnectDB(t *testing.T) {
	t.Setenv("APP_ENV", "DEV")
	t.Setenv("OPENBAO_TOKEN", "token")
	cfg := config.Read()
	db, dbErr := ConnectDB(cfg, cfg.Test.DbPosition)
	Ok(t, dbErr)
	t.Cleanup(func() { db.Close() })
}
