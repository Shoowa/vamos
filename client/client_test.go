package client_test

import (
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/Shoowa/vamos/client"
	"github.com/Shoowa/vamos/config"
	"github.com/Shoowa/vamos/logging"
	. "github.com/Shoowa/vamos/testhelper"

	"golang.org/x/sync/errgroup"
)

const GOOGLE = "http://www.google.com"

var (
	webclient *http.Client
	attempts  int
)

func TestMain(m *testing.M) {
	os.Setenv("VAMOS_CONFIG", "../config/dev.json")
	os.Setenv("OPENBAO_TOKEN", "token")

	// Erect custom http.Client before tests run. Re-define rate limiting so the
	// tests aren't influenced by the config file.
	cfg := config.Read()
	cfg.Client.RateLimiter.Active = true
	cfg.Client.RateLimiter.Window = 5
	cfg.Client.RateLimiter.Amount = 5
	cfg.Client.RateLimiter.Burst = 5
	attempts = int(cfg.Client.RateLimiter.Amount) * 2
	logger := logging.CreateLogger(cfg)
	webclient = client.CreateHTTPClient(cfg, logger)

	// Execute tests.
	code := m.Run()

	os.Unsetenv("VAMOS_CONFIG")
	os.Unsetenv("OPENBAO_TOKEN")

	// Report results.
	os.Exit(code)
}

func Test_RateLimit(t *testing.T) {
	errGroup, errGroupCtx := errgroup.WithContext(t.Context())
	req, _ := http.NewRequestWithContext(errGroupCtx, http.MethodGet, GOOGLE, nil)

	for range attempts {
		errGroup.Go(func() error {
			resp, respErr := webclient.Do(req)
			if respErr != nil {
				return respErr
			}
			NotEqual(t, resp.StatusCode, http.StatusTooManyRequests)

			// drain connection.
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			return nil
		})
	}

	err := errGroup.Wait()
	Ok(t, err)
}
