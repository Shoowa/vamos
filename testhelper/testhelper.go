package testhelper

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Shoowa/vamos/config"
	"github.com/Shoowa/vamos/data/rdbms"
	"github.com/Shoowa/vamos/router"
	"github.com/Shoowa/vamos/secrets"

	"github.com/jackc/pgx/v5"
)

const (
	TIMEOUT_REQ = time.Second * 1
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

// DesiredError fails the test when an expected error is missing.
func DesiredError(t *testing.T, err error, actual any) {
	t.Helper()

	if err == nil {
		t.Errorf("expected an error; actual: %#v", actual)
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

func ErrorIs(t *testing.T, actual, expectation error) {
	t.Helper()

	if !errors.Is(actual, expectation) {
		t.Errorf("actual: %v; expectation: %v", actual, expectation)
	}
}

func ErrorAs(t *testing.T, actual error, target any) {
	t.Helper()

	if actual == nil {
		t.Errorf("actual: NIL; expectation: assignable to: %T", target)
		return
	}

	if !errors.As(actual, target) {
		t.Errorf("actual: %v; expectation assignable to: %T", actual, target)
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

// DbHandle is a convenient struct that represents a single connection to a
// Postgres server. It includes required data for opening and closing that
// connection.
type DbHandle struct {
	*pgx.Conn
	timer time.Duration
	ctx   context.Context
}

// SetTimer defines the time that a handle will wait to execute a SQL statement.
func (db *DbHandle) SetTimer(seconds int) {
	db.timer = time.Second * time.Duration(seconds)
}

// CloseConnection simply terminates a TCP connection to a Postgres server.
func (db *DbHandle) CloseConnection() {
	closeTimer, cancelClose := context.WithTimeout(db.ctx, db.timer)
	defer cancelClose()
	db.Close(closeTimer)
}

// ConnectDB creates a Database handle, and is meant to be executed inside the
// TestMain func. It provides a single connection to a Postgres server
// independent of the application's connect pool. This connection can be re-used
// in multiple tests. This connection can be closed after the test runs inside
// the TestMain func.
func ConnectDB() *DbHandle {
	cfg := config.Read()
	dbConfig := rdbms.WhichDB(cfg, cfg.Test.DbPosition)
	credString, credErr := rdbms.Credentials(dbConfig)
	if credErr != nil {
		panic(credErr)
	}

	connConfig, configErr := pgx.ParseConfig(credString)
	if configErr != nil {
		panic(configErr)
	}

	secReader := new(secrets.SkeletonKey)
	secReader.Create(cfg)
	pw, pwErr := secReader.ReadPathAndKey(dbConfig.Secret, dbConfig.SecretKey)
	if pwErr != nil {
		panic(pwErr)
	}

	connConfig.Password = pw

	// Add TLS config to single Postgres connection.
	if dbConfig.Sslmode == true {
		tlsInfo, tlsErr := secReader.ConfigureTLSwithCA(cfg.HttpServer)
		if tlsErr != nil {
			panic(tlsErr)
		}
		tlsInfo.ServerName = dbConfig.Host
		connConfig.Config.TLSConfig = tlsInfo
	}

	// Set timer for opening a connection.
	ctx := context.Background()
	openTimer, cancelOpen := context.WithTimeout(ctx, TIMEOUT_REQ)
	defer cancelOpen()

	// Use a single connection. This isn't a pool.
	db, connErr := pgx.ConnectConfig(openTimer, connConfig)
	if connErr != nil {
		panic(connErr)
	}

	return &DbHandle{
		db,
		time.Second * time.Duration(1),
		ctx,
	}
}

// WriteAndDeleteData inserts and erases data from Postgres. Useful for setting
// up a test.
func (db *DbHandle) WriteAndDeleteData(t *testing.T, fileWrite, fileDelete string) {
	fakeData, fileErr := os.ReadFile(fileWrite)
	if fileErr != nil {
		t.Fatal(fileErr)
	}

	// Set timer for issuing SQL command that writes data into the table.
	cmdTimer, cancelCommand := context.WithTimeout(db.ctx, db.timer)
	defer cancelCommand()

	// Write fake data to Postgres.
	_, writeErr := db.Exec(cmdTimer, string(fakeData))
	if writeErr != nil {
		closeTimer, cancelClose := context.WithTimeout(db.ctx, db.timer)
		defer cancelClose()
		db.Close(closeTimer)
		t.Fatal(writeErr)
	}

	// Ensure deletion occurs.
	t.Cleanup(func() {
		eraseData, fileErr := os.ReadFile(fileDelete)
		if fileErr != nil {
			t.Fatal(fileErr)
		}

		// Erase data from Postgres.
		cmdTimer, cancelCommand := context.WithTimeout(db.ctx, db.timer)
		defer cancelCommand()
		_, delErr := db.Exec(cmdTimer, string(eraseData))
		if delErr != nil {
			t.Fatal(delErr)
		}
	})
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
