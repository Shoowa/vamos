//go:build integration

package cache_test

import (
	"os"
	"testing"

	"github.com/Shoowa/vamos/config"
	. "github.com/Shoowa/vamos/data/cache"
	"github.com/Shoowa/vamos/secrets"
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

func Test_ConnectCache(t *testing.T) {
	sk := new(secrets.SkeletonKey)
	sk.Create(cfg)

	cache, cErr := CreateClient(cfg, sk)
	Ok(t, cErr)

	ctx := t.Context()
	pong, pErr := cache.Ping(ctx).Result()
	Ok(t, pErr)

	Equals(t, "PONG", pong)

	t.Cleanup(func() { cache.Close() })
}
