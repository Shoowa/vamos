package testhelper

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"vamos/internal/config"
	"vamos/internal/data/rdbms"
	"vamos/internal/server"
)

const (
	DB_FIRST = 0
	PROJECT  = "vamos"
)

// assert fails the test if the condition is false.
func Assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// ok fails the test if an err is not nil.
func Ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func Equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}

func Change_to_project_root() {
	wd, _ := os.Getwd()
	for !strings.HasSuffix(wd, PROJECT) {
		wd = filepath.Dir(wd)
	}
	changeErr := os.Chdir(wd)
	if changeErr != nil {
		panic(changeErr)
	}
}

func createRouter(t *testing.T) *server.Bundle {
	cfg := config.Read()
	logger := slog.New(slog.DiscardHandler)

	db1, db1Err := rdbms.ConnectDB(cfg, DB_FIRST)
	if db1Err != nil {
		logger.Error(db1Err.Error())
		panic(db1Err)
	}

	backbone := server.NewBackbone(
		server.WithLogger(logger),
		server.WithQueryHandleForFirstDB(db1),
		server.WithDbHandle(db1),
	)

	router := server.NewRouter(backbone)

	return router
}

type testServer struct {
	*httptest.Server
}

func CreateTestServer(t *testing.T) *testServer {
	jar, jErr := cookiejar.New(nil)
	if jErr != nil {
		t.Fatal(jErr)
	}

	router := createRouter(t)
	s := httptest.NewServer(router)

	// Enable saving response cookies for subsequent requests.
	s.Client().Jar = jar

	// Disable redirect to see the first response.
	s.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{s}
}

func (tsrv *testServer) Get(t *testing.T, path string) (int, http.Header, string) {
	client := tsrv.Client()
	r, rErr := client.Get(tsrv.URL + path)
	if rErr != nil {
		t.Fatal(rErr)
	}

	defer r.Body.Close()
	body, bodyErr := io.ReadAll(r.Body)
	if bodyErr != nil {
		t.Fatal(bodyErr)
	}

	body = bytes.TrimSpace(body)
	return r.StatusCode, r.Header, string(body)
}
