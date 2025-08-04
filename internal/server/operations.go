package server

import (
	"context"
	"runtime"
	"runtime/pprof"
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
	Heap  bool
}

func (h *Health) PassFail() bool {
	return h.Rdbms && h.Heap
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
	heapTimer := time.Duration(cfg.Health.HeapTimer)

	// Use closure to configure method CheckHeapSize.
	heapSize := 1024 * 1024 * cfg.Health.HeapSize
	checkHeapSize := func() { b.CheckHeapSize(heapSize) }

	go beep(pingDbTimer, b.PingDB)
	go beep(heapTimer, checkHeapSize)
}

func (b *Backbone) Write(p []byte) (n int, err error) {
	b.HeapSnapshot.Reset()
	return b.HeapSnapshot.Write(p)
}

func (b *Backbone) CheckHeapSize(threshold uint64) {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	if stats.HeapAlloc < threshold {
		b.Health.Heap = true
		return
	}

	b.Health.Heap = false
	b.Logger.Warn("Heap surpassed threshold!", "threshold", threshold, "allocated", stats.HeapAlloc)
	err := pprof.WriteHeapProfile(b)
	if err != nil {
		b.Logger.Error("Error writing heap profile", "ERR:", err.Error())
	}
}
