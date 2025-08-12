//go:build integration

package rdbms_test

import (
	"context"
	"os"
	"testing"
	"time"

	"vamos/internal/config"
	. "vamos/internal/data/rdbms"
	. "vamos/internal/testhelper"
	"vamos/sqlc/data/first"

	"github.com/jackc/pgx/v5"
)

const (
	TIMEOUT_TEST = time.Second * 1
	TIMEOUT_READ = time.Second * 3
	TEST_DB_POS  = 0
	FAKE_DATA    = "_testdata/fake_data_db1.sql"
	AUTHOR       = "Chaucer"
)

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
	Change_to_project_root()
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

// First, test opening a connection to a database.
func Test_ConnectDB(t *testing.T) {
	t.Setenv("APP_ENV", "DEV")
	t.Setenv("OPENBAO_TOKEN", "token")
	db, dbErr := ConnectDB(config.Read(), TEST_DB_POS)
	Ok(t, dbErr)
	t.Cleanup(func() { db.Close() })
}

// Second, test reading data concurrently.
func Test_ReadingData(t *testing.T) {
	t.Setenv("APP_ENV", "DEV")
	t.Setenv("OPENBAO_TOKEN", "token")

	db, _ := ConnectDB(config.Read(), TEST_DB_POS)
	q := FirstDB_AdoptQueries(db)

	timer, _ := context.WithTimeout(context.Background(), TIMEOUT_READ)

	t.Run("Read one author", func(t *testing.T) {
		readOneAuthor(t, q, timer)
	})

	t.Run("Read many authors", func(t *testing.T) {
		readManyAuthors(t, q, timer)
	})

	t.Run("Read most productive author", func(t *testing.T) {
		readMostProductiveAuthor(t, q, timer)
	})

	t.Run("Read most productive author & book", func(t *testing.T) {
		readMostProductiveAuthorAndBook(t, q, timer)
	})

	t.Cleanup(func() { db.Close() })
}

func readOneAuthor(t *testing.T, q *first.Queries, ctx context.Context) {
	t.Parallel()
	timer, cancel := context.WithTimeout(ctx, TIMEOUT_TEST)
	defer cancel()
	result, err := q.GetAuthor(timer, AUTHOR)

	Ok(t, err)
	Equals(t, result.Name, AUTHOR)
}

func readManyAuthors(t *testing.T, q *first.Queries, ctx context.Context) {
	t.Parallel()
	timer, cancel := context.WithTimeout(ctx, TIMEOUT_TEST)
	defer cancel()
	result, err := q.ListAuthors(timer)

	Ok(t, err)
	Equals(t, 5, len(result))
}

func readMostProductiveAuthor(t *testing.T, q *first.Queries, ctx context.Context) {
	t.Parallel()
	timer, cancel := context.WithTimeout(ctx, TIMEOUT_TEST)
	defer cancel()
	result, err := q.MostProductiveAuthor(timer)

	Ok(t, err)
	Equals(t, AUTHOR, result)
}

func readMostProductiveAuthorAndBook(t *testing.T, q *first.Queries, ctx context.Context) {
	t.Parallel()
	timer, cancel := context.WithTimeout(ctx, TIMEOUT_TEST)
	defer cancel()
	result, err := q.MostProductiveAuthorAndBook(timer)

	Ok(t, err)
	Equals(t, AUTHOR, result.Author.Name)
	Equals(t, "The Canterbury Tales", result.Book.Title)
}
