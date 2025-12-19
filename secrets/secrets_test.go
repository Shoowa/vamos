package secrets_test

import (
	"os"
	"strings"
	"testing"

	"github.com/Shoowa/vamos/config"
	. "github.com/Shoowa/vamos/secrets"
	. "github.com/Shoowa/vamos/testhelper"
)

func TestMain(m *testing.M) {
	Change_to_project_root()
	code := m.Run()
	os.Exit(code)
}

func Test_Configuration(t *testing.T) {
	t.Setenv("APP_ENV", "DEV")
	t.Setenv("OPENBAO_TOKEN", "token")

	cfg := config.Read()
	o := cfg.Secrets.Openbao
	o.ReadConfig()
	config := ReadConfig(cfg)
	o.ReadToken()

	Assert(t, strings.Contains(config.Address, "localhost"), "URL misconfigured")
	Equals(t, "token", o.Token)
}
