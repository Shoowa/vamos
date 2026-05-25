package testhelper

import (
	"bytes"
	"context"
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
	"time"

	"github.com/Shoowa/vamos/config"
	"github.com/Shoowa/vamos/data/rdbms"
	"github.com/Shoowa/vamos/router"

	"github.com/jackc/pgx/v5"
)

const (
	TIMEOUT = time.Second * 1
)

func True(t *testing.T, actual bool) {
	t.Helper()

	if !actual {
		t.Errorf("actual: %v; expectation: TRUE", actual)
	}
}

func Ok(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Errorf("Error detected: %s", err)
	}
}

func Equal[T any](t *testing.T, actual, expectation T) {
	t.Helper()

	if !isEqual(actual, expectation) {
		t.Errorf("actual: %v; expectation: %v", actual, expectation)
	}
}

func isEqual[T any](actual, expectation T) bool {
	// Are both values actually nil?
	if isNil(actual) && isNil(expectation) {
		return true
	}

	// Are they truly the same?
	return reflect.DeepEqual(actual, expectation)
}

func isNil(v any) bool {
	if v == nil {
		return true
	}

	// Inspect underlying type of v, and return true when it is nullable, e.g.,
	// pointer, slice, map, with a value of nil.
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		return rv.IsNil()
	}

	// Other types like bool, int, string are never nil.
	return false
}

func NotEqual[T any](t *testing.T, actual, expectation T) {
	t.Helper()

	if isEqual(actual, expectation) {
		t.Errorf("actual: %v; expected values to be different", actual)
	}
}

func Nil(t *testing.T, actual any) {
	t.Helper()

	if !isNil(actual) {
		t.Errorf("actual: %v; expectation: NIL", actual)
	}
}

func NotNil(t *testing.T, actual any) {
	t.Helper()

	if isNil(actual) {
		t.Errorf("actual: NIL; expectation: NOT NIL")
	}
}

// Change_to_project_root repositions a test executable's starting point so it
// can find the config file.
func Change_to_project_root() {
	dirName := os.Getenv("PROJECT_NAME")
	wd, _ := os.Getwd()
	for !strings.HasSuffix(wd, dirName) {
		wd = filepath.Dir(wd)
	}
	changeErr := os.Chdir(wd)
	if changeErr != nil {
		panic(changeErr)
	}
}

// CreateTestTable writes data into a development Postgres server.
func CreateTestTable(timer context.Context) error {
	cfg := config.Read()
	dbConfig := rdbms.WhichDB(cfg, cfg.Test.DbPosition)
	credString, credErr := rdbms.Credentials(dbConfig)
	if credErr != nil {
		return credErr
	}

	// Set timer for opening a connection.
	openTimer, cancelOpen := context.WithTimeout(timer, TIMEOUT)
	defer cancelOpen()

	// Use a single connection. This isn't a pool.
	db, connErr := pgx.Connect(openTimer, credString)
	if connErr != nil {
		return connErr
	}

	// Set timer for closing the connection.
	closeTimer, cancelClose := context.WithTimeout(timer, TIMEOUT)
	defer cancelClose()
	defer db.Close(closeTimer)

	// Set timer for issuing SQL command that writes data into the table.
	cmdTimer, cancelCommand := context.WithTimeout(timer, time.Second*3)
	defer cancelCommand()

	fakeData, fileErr := os.ReadFile(cfg.Test.FakeData)
	if fileErr != nil {
		return fileErr
	}
	db.Exec(cmdTimer, string(fakeData))
	return nil
}

type testServer struct {
	*httptest.Server
}

// CreateTestServer conveniently creates a configured server for testing routes.
func CreateTestServer(t *testing.T) *testServer {
	cfg := config.Read()
	logger := slog.New(slog.DiscardHandler)

	db1, db1Err := rdbms.ConnectDB(cfg, cfg.Test.DbPosition)
	if db1Err != nil {
		logger.Error(db1Err.Error())
		panic(db1Err)
	}

	backbone := router.NewBackbone(
		router.WithLogger(logger),
		router.WithDbHandle(db1),
	)

	router := router.NewRouter(cfg, backbone)
	s := httptest.NewServer(router)

	jar, jErr := cookiejar.New(nil)
	if jErr != nil {
		t.Fatal(jErr)
	}

	// Enable saving response cookies for subsequent requests.
	s.Client().Jar = jar

	// Disable redirect to see the first response.
	s.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{s}
}

// Get is a method of the testServer that conveniently creates a client and
// reads a response and reports the HTTP status, Headers, and body.
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

// CreateTestServerExtDeps conveniently creates a configured server in a
// downstream test executable. This is one of the two reasons the Gatherer
// interface was created. To easily test routes with dependencies in a
// downstream executable.
func CreateTestServerExtDeps(t *testing.T, d router.Gatherer) *testServer {
	cfg := config.Read()
	logger := slog.New(slog.DiscardHandler)

	db1, db1Err := rdbms.ConnectDB(cfg, cfg.Test.DbPosition)
	if db1Err != nil {
		logger.Error(db1Err.Error())
		panic(db1Err)
	}

	backbone := router.NewBackbone(
		router.WithLogger(logger),
		router.WithDbHandle(db1),
	)

	// Incorporate downstream HTTP Handlers into this upstream test server.
	d.AddBackbone(backbone)
	router := router.NewRouter(cfg, d)

	s := httptest.NewServer(router)

	jar, jErr := cookiejar.New(nil)
	if jErr != nil {
		t.Fatal(jErr)
	}

	// Enable saving response cookies for subsequent requests.
	s.Client().Jar = jar

	// Disable redirect to see the first response.
	s.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{s}
}
