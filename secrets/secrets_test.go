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
