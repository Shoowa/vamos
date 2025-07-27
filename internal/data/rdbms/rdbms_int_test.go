//go:build integration

package rdbms_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"vamos/internal/config"
	. "vamos/internal/data/rdbms"
	. "vamos/internal/testhelper"
)

const (
	TIMEOUT_TEST = time.Second * 2
	PROJECT      = "vamos"
	TEST_DB_POS  = 0
	TEST_USER    = "tester"
	TEST_DB      = "test_data"
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

func TestMain(m *testing.M) {
	change_to_project_root()
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
