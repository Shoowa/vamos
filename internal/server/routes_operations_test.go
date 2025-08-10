//go:build !integration

package server_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	. "vamos/internal/server"
	. "vamos/internal/testhelper"
)

func Test_Healthcheck_Initial_One_Resource_Down(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	b := NewBackbone(WithLogger(logger))

	rec := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, "/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	b.Healthcheck(rec, r)

	result := rec.Result()
	Equals(t, http.StatusServiceUnavailable, result.StatusCode)
}
