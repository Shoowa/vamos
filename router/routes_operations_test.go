//go:build integration

package router_test

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Shoowa/vamos/config"
	"github.com/Shoowa/vamos/data/cache"
	"github.com/Shoowa/vamos/data/rdbms"
	"github.com/Shoowa/vamos/metrics"
	"github.com/Shoowa/vamos/router"
	"github.com/Shoowa/vamos/secrets"
	. "github.com/Shoowa/vamos/testhelper"
)

const (
	DIR_ABOVE = "../"
)

var (
	server *testServer
	client *testClient
)

type testServer struct {
	*httptest.Server
}

func createTestServer(cfg *config.Config) *testServer {
	logger := slog.New(slog.DiscardHandler)
	db1, db1Err := rdbms.ConnectDB(cfg, cfg.Test.DbPosition)
	if db1Err != nil {
		panic(db1Err.Error())
	}

	secretsReader := new(secrets.SkeletonKey)
	secretsReader.Create(cfg)

	cache, cacheErr := cache.CreateClient(cfg, secretsReader)
	if cacheErr != nil {
		panic(cacheErr.Error())
	}

	backbone := router.NewBackbone(
		router.WithLogger(logger),
		router.WithDbHandle(db1),
		router.WithCache(cache),
	)

	rtr := router.NewRouter(cfg, backbone)
	server := httptest.NewServer(rtr)

	jar, jErr := cookiejar.New(nil)
	if jErr != nil {
		panic(jErr.Error())
	}

	// Enable saving response cookies for subsequent requests.
	server.Client().Jar = jar

	// Disable redirect to see the first response.
	server.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{server}
}

func (ts *testServer) createClient() *testClient {
	return &testClient{
		ts.Client(),
		ts.URL,
	}
}

type testClient struct {
	*http.Client
	URL string
}

type testResponse struct {
	status int
	header http.Header
	body   string
}

func (tc *testClient) GET(t *testing.T, path string) testResponse {
	resp, rErr := tc.Get(tc.URL + path)
	if rErr != nil {
		t.Fatal(rErr)
	}
	body, bErr := io.ReadAll(resp.Body)
	if bErr != nil {
		t.Fatal(bErr)
	}
	body = bytes.TrimSpace(body)
	return testResponse{
		status: resp.StatusCode,
		header: resp.Header,
		body:   string(body),
	}
}

func TestMain(m *testing.M) {
	os.Setenv("VAMOS_CONFIG", "../config/dev.json")
	os.Setenv("OPENBAO_TOKEN", "token")

	cfg := config.Read()
	metrics.Registry = metrics.CreateLoadedRegistry(cfg.Metrics)
	// test executable needs a different relative path
	tls := cfg.Secrets.Openbao.TlsClient
	tls.CaPath = DIR_ABOVE + tls.CaPath
	tls.CertPath = DIR_ABOVE + tls.CertPath
	tls.KeyPath = DIR_ABOVE + tls.KeyPath

	server = createTestServer(cfg)
	client = server.createClient()

	code := m.Run()

	server.Close()
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
