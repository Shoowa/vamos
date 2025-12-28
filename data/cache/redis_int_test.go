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

func TestMain(m *testing.M) {
	os.Setenv("APP_ENV", "DEV")
	Change_to_project_root()
	os.Unsetenv("APP_ENV")

	code := m.Run()
	os.Exit(code)
}

func Test_ConnectCache(t *testing.T) {
	t.Setenv("APP_ENV", "DEV")
	t.Setenv("OPENBAO_TOKEN", "token")

	cfg := config.Read()
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
