package main

import (
	"vamos/internal/config"
	"vamos/internal/logging"
)

func main() {
	cfg := config.Read()

	logging.CreateLogger(cfg)
}
