//go:build integration

package rdbms_test

import (
	"os"
	"testing"

	"github.com/Shoowa/vamos/config"
	"github.com/Shoowa/vamos/data/rdbms"
	. "github.com/Shoowa/vamos/testhelper"
)

const (
	DIR_ABOVE = "../../"
)

var (
	cfg *config.Config
)

func TestMain(m *testing.M) {
	os.Setenv("VAMOS_CONFIG", "../../config/dev.json")
	os.Setenv("OPENBAO_TOKEN", "token")

	cfg = config.Read()

	// test executable needs a different relative path
	tls := cfg.Secrets.Openbao.TlsClient
	tls.CaPath = DIR_ABOVE + tls.CaPath
	tls.CertPath = DIR_ABOVE + tls.CertPath
	tls.KeyPath = DIR_ABOVE + tls.KeyPath

	code := m.Run()

	os.Unsetenv("VAMOS_CONFIG")
	os.Unsetenv("OPENBAO_TOKEN")
	os.Exit(code)
}

// First, test opening a connection to a database.
func Test_ConnectDB(t *testing.T) {
	db, dbErr := rdbms.ConnectDB(cfg, cfg.Test.DbPosition)
	Ok(t, dbErr)
	t.Cleanup(func() { db.Close() })
}
