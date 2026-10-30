//go:build !integration

package router_test

import (
	"net/http"
	"os"
	"testing"

	. "github.com/Shoowa/vamos/testhelper"
)

func TestMain(m *testing.M) {
	os.Setenv("VAMOS_CONFIG", "../config/dev.json")
	os.Setenv("OPENBAO_TOKEN", "token")

	code := m.Run()

	os.Unsetenv("VAMOS_CONFIG")
	os.Unsetenv("OPENBAO_TOKEN")
	os.Exit(code)
}

func Test_Healthcheck_Initial_One_Resource_Down(t *testing.T) {
	srv := CreateTestServer(t)

	code, _, _ := srv.Get(t, "/health")
	Equal(t, http.StatusServiceUnavailable, code)
	t.Cleanup(func() { srv.Close() })
}
