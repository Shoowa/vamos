package main

import (
	"vamos/internal/config"
	"vamos/internal/data/rdbms"
	"vamos/internal/logging"
)

const (
	DB_FIRST = 0
)

func main() {
	cfg := config.Read()

	logger := logging.CreateLogger(cfg)

	db1, db1Err := rdbms.ConnectDB(cfg, DB_FIRST)
	if db1Err != nil {
		logger.Error(db1Err.Error())
		panic(db1Err)
	}
	defer db1.Close()

	rdbms.FirstDB_AdoptQueries(db1)
}
