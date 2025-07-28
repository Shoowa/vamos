//go:build integration

package rdbms_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"vamos/internal/config"
	. "vamos/internal/data/rdbms"
	. "vamos/internal/testhelper"

	"github.com/jackc/pgx/v5"
)

const (
	TIMEOUT_TEST = time.Second * 1
	PROJECT      = "vamos"
	TEST_DB_POS  = 0
	FAKE_DATA    = "_testdata/fake_data_db1.sql"
)

func change_to_project_root() {
	wd, _ := os.Getwd()
	for !strings.HasSuffix(wd, PROJECT) {
		wd = filepath.Dir(wd)
	}
	changeErr := os.Chdir(wd)
	if changeErr != nil {
		panic(changeErr.Error())
	}
}

func createTestTable(timer context.Context) error {
	// Read configuration information to establish connection.
	dbConfig := WhichDB(config.Read(), TEST_DB_POS)
	credString, credErr := Credentials(dbConfig)
	if credErr != nil {
		return credErr
	}

	// Set timer for opening a connection.
	openTimer, cancelOpen := context.WithTimeout(timer, TIMEOUT_TEST)
	defer cancelOpen()

	// Use a single connection. This isn't a pool.
	db, connErr := pgx.Connect(openTimer, credString)
	if connErr != nil {
		return connErr
	}

	// Set timer for closing the connection.
	closeTimer, cancelClose := context.WithTimeout(timer, TIMEOUT_TEST)
	defer cancelClose()
	defer db.Close(closeTimer)

	// Set timer for issuing SQL command that writes data into the table.
	cmdTimer, cancelCommand := context.WithTimeout(timer, time.Second*3)
	defer cancelCommand()

	fakeData, fileErr := os.ReadFile(FAKE_DATA)
	if fileErr != nil {
		return fileErr
	}
	db.Exec(cmdTimer, string(fakeData))
	return nil
}

func TestMain(m *testing.M) {
	os.Setenv("APP_ENV", "DEV")
	change_to_project_root()
	timer, _ := context.WithTimeout(context.Background(), time.Second*5)

	// Setup common resource for all integration tests in only this package.
	dbErr := createTestTable(timer)
	if dbErr != nil {
		panic(dbErr)
	}
	os.Unsetenv("APP_ENV")

	code := m.Run()
	os.Exit(code)
}

func Test_ConnectDB(t *testing.T) {
	t.Setenv("APP_ENV", "DEV")
	t.Setenv("OPENBAO_TOKEN", "token")
	db, dbErr := ConnectDB(config.Read(), TEST_DB_POS)
	defer db.Close()
	Ok(t, dbErr)
}
