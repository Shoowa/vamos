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

const (
	TIMEOUT_TEST = time.Second * 1
	TIMEOUT_READ = time.Second * 3
	AUTHOR       = "Chaucer"
)

func TestMain(m *testing.M) {
	os.Setenv("APP_ENV", "DEV")
	Change_to_project_root()
	timer, _ := context.WithTimeout(context.Background(), time.Second*5)

	// Setup common resource for all integration tests in only this package.
	dbErr := CreateTestTable(timer)
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
	cfg := config.Read()
	db, dbErr := rdbms.ConnectDB(cfg, cfg.Test.DbPosition)
	Ok(t, dbErr)
	t.Cleanup(func() { db.Close() })
}

// Second, test reading data concurrently.
func Test_ReadingData(t *testing.T) {
	t.Setenv("APP_ENV", "DEV")
	t.Setenv("OPENBAO_TOKEN", "token")

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
