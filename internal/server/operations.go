package server

import (
	"context"
	"time"

	"vamos/internal/config"
)

const TIMEOUT_PING = time.Millisecond * 299

// beep performs a task every X seconds.
func beep(seconds time.Duration, task func()) {
	ticker := time.NewTicker(seconds * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			task()
		}
	}
}

// Health offers summarized data.
type Health struct {
	Rdbms bool
}

func (h *Health) PassFail() bool {
	return h.Rdbms
}

func (b *Backbone) PingDB() {
	timer, cancel := context.WithTimeout(context.Background(), TIMEOUT_PING)
	defer cancel()

	err := b.DbHandle.Ping(timer)
	if err != nil {
		b.Health.Rdbms = false
		b.Logger.Error("Failed ping", "Rdbms", err.Error())
		return
	}
	b.Health.Rdbms = true
}

func (b *Backbone) SetupHealthChecks(cfg *config.Config) {
	// Report status of connection upon ignition.
	b.PingDB()

	pingDbTimer := time.Duration(cfg.Health.PingDbTimer)
	go beep(pingDbTimer, b.PingDB)
}
