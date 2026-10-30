//go:build integration

package data_test

import (
	"context"
	"os"
	"testing"
	"time"

	"_example/sqlc/data/first"

	"github.com/Shoowa/vamos/config"
	"github.com/Shoowa/vamos/data/rdbms"
	. "github.com/Shoowa/vamos/testhelper"
)

// Define common values, like HTTP Headers and common data.
const (
	TIMEOUT_TEST = time.Second * 1
	TIMEOUT_READ = time.Second * 3
	AUTHOR       = "Chaucer"
)

// Define a package variable that can be used during the setup of the test
// suite, and in each test. This is a single connection to a Postgres server,
// unlike the application's connection pool.
var (
	testDB *DbHandle
)

func TestMain(m *testing.M) {
	// Enable reading the configuration file.
	os.Setenv("VAMOS_CONFIG", "../config/dev.json")

	// Enable reading credentials from the Openbao server.
	os.Setenv("OPENBAO_TOKEN", "token")

	// Connect to the local database before any tests run. Notice this uses the
	// package variable defined above. It will be re-used in the tests.
	testDB = ConnectDB()
	testDB.SetTimer(2)

	code := m.Run()

	testDB.CloseConnection()
	os.Unsetenv("VAMOS_CONFIG")
	os.Unsetenv("OPENBAO_TOKEN")
	os.Exit(code)
}

// First, test opening a connection to a database.
func Test_ConnectDB(t *testing.T) {
	cfg := config.Read()
	db, dbErr := rdbms.ConnectDB(cfg, cfg.Test.DbPosition)
	Ok(t, dbErr)
	t.Cleanup(func() { db.Close() })
}

// Second, test reading data concurrently. Notice it uses two different database
// handles. The first is the single connection that adds data to the table. The
// second is the connection pool associated with the application that the HTTP
// Handlers will use.
func Test_ReadingData(t *testing.T) {
	testDB.WriteAndDeleteData(t, "add_table_data.sql", "erase_table_data.sql")

	cfg := config.Read()
	db, _ := rdbms.ConnectDB(cfg, cfg.Test.DbPosition)
	q := first.New(db)

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
	Equal(t, result.Name, AUTHOR)
}

func readManyAuthors(t *testing.T, q *first.Queries, ctx context.Context) {
	t.Parallel()
	timer, cancel := context.WithTimeout(ctx, TIMEOUT_TEST)
	defer cancel()
	result, err := q.ListAuthors(timer)

	Ok(t, err)
	Equal(t, 5, len(result))
}

func readMostProductiveAuthor(t *testing.T, q *first.Queries, ctx context.Context) {
	t.Parallel()
	timer, cancel := context.WithTimeout(ctx, TIMEOUT_TEST)
	defer cancel()
	result, err := q.MostProductiveAuthor(timer)

	Ok(t, err)
	Equal(t, AUTHOR, result)
}

func readMostProductiveAuthorAndBook(t *testing.T, q *first.Queries, ctx context.Context) {
	t.Parallel()
	timer, cancel := context.WithTimeout(ctx, TIMEOUT_TEST)
	defer cancel()
	result, err := q.MostProductiveAuthorAndBook(timer)

	Ok(t, err)
	Equal(t, AUTHOR, result.Author.Name)
	Equal(t, "The Canterbury Tales", result.Book.Title)
}
