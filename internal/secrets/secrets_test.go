package secrets_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"vamos/internal/config"
	. "vamos/internal/secrets"
	. "vamos/internal/testhelper"
)

// Reposition test executable root to read config
func change_to_project_root() {
	wd, _ := os.Getwd()
	for !strings.HasSuffix(wd, "vamos") {
		wd = filepath.Dir(wd)
	}
	changeErr := os.Chdir(wd)
	if changeErr != nil {
		panic(changeErr)
	}
}

func TestMain(m *testing.M) {
	change_to_project_root()
	code := m.Run()
	os.Exit(code)
}

func Test_Configuration(t *testing.T) {
	t.Setenv("APP_ENV", "DEV")
	t.Setenv("OPENBAO_TOKEN", "token")

	cfg := config.Read()
	o := cfg.Secrets.Openbao
	o.ReadConfig()
	client := ReadConfig(cfg)

	Assert(t, strings.Contains(client.Address, "localhost"), "URL misconfigured")
	Equals(t, "token", o.Token)
}
