package secrets_test

import (
	"os"
	"strings"
	"testing"

	"github.com/Shoowa/vamos/config"
	. "github.com/Shoowa/vamos/secrets"
	. "github.com/Shoowa/vamos/testhelper"
)

const (
	DIR_ABOVE = "../"
)

var (
	cfg *config.Config
)

func TestMain(m *testing.M) {
	os.Setenv("VAMOS_CONFIG", "../config/dev.json")
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

func Test_SkeletonKeyOpenbao(t *testing.T) {
	sk := new(SkeletonKey)
	sk.Create(cfg)

	addr := sk.Openbao.Address()
	tok := sk.Openbao.Token()
	Assert(t, strings.Contains(addr, "localhost"), "Lacks localhost in host address.")
	Assert(t, strings.Contains(tok, "token"), "Lacks token from environment.")
}
