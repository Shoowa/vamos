//go:build !integration

package server_test

import (
	"net/http"
	"os"
	"testing"

	. "vamos/internal/testhelper"
)

func TestMain(m *testing.M) {
	Change_to_project_root()
	code := m.Run()
	os.Exit(code)
}

func Test_Healthcheck_Initial_One_Resource_Down(t *testing.T) {
	t.Setenv("APP_ENV", "DEV")
	t.Setenv("OPENBAO_TOKEN", "token")

	srv := CreateTestServer(t)

	code, _, _ := srv.Get(t, "/health")
	Equals(t, http.StatusServiceUnavailable, code)
	t.Cleanup(func() { srv.Close() })
}
