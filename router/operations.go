package router

import (
	"context"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/Shoowa/vamos/config"
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

// Health offers summarized data that can be read on the /health endpoint. No
// matter how often an external service hammers the /health endpoint, the
// application will easily read a boolean to reply. The real work of evaluating
// the health of the application and dependencies is left to background
// goroutines.
type Health struct {
	Rdbms    bool
	Heap     bool
	Routines bool
}

// PassFail evaluates the totality of dependencies and the application.
func (h *Health) PassFail() bool {
	return h.Rdbms && h.Heap && h.Routines
}

// Ping evaluates the ability to contact a Postgres server
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

// SetupHealthChecks reads configured values, and leverages closures to apply
// them to Backbone methods designed to run periodically. Each of those methods
// evaluates one condition in the application or a dependency.
func (b *Backbone) SetupHealthChecks(cfg *config.Config) {
	// Report status of connection upon ignition.
	b.PingDB()

	pingDbTimer := time.Duration(cfg.Health.PingDbTimer)
	heapTimer := time.Duration(cfg.Health.HeapTimer)
	routinesTimer := time.Duration(cfg.Health.RoutTimer)

	// Use closure to configure method CheckHeapSize.
	heapSize := 1024 * 1024 * cfg.Health.HeapSize
	checkHeapSize := func() { b.CheckHeapSize(heapSize) }

	// Use closure to configure method CheckNumRoutines.
	limit := runtime.NumCPU() * cfg.Health.RoutinesPerCore
	checkNumRoutines := func() { b.CheckNumRoutines(limit) }

	go beep(pingDbTimer, b.PingDB)
	go beep(heapTimer, checkHeapSize)
	go beep(routinesTimer, checkNumRoutines)
}

// Write is a method on the Backbone struct designed to conform to an interface,
// so that the Backbone can easily write runtime information to its Buffer
// field.
func (b *Backbone) Write(p []byte) (n int, err error) {
	b.HeapSnapshot.Reset()
	return b.HeapSnapshot.Write(p)
}

// CheckHeapSize reads the runtime to determine whether or not the size of the
// Heap surprasses a configured value. When the heap exceeds a configured size,
// a health check will fail, and Heap data will be written to a buffer. That
// data is sensitive, and must be exfiltrated by another function.
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

// CheckNumRoutines reads the runtime to determine whether or not the amount of
// routines surpasses a configured value, then fail a health check.
func (b *Backbone) CheckNumRoutines(limit int) {
	amount := runtime.NumGoroutine()
	if amount < limit {
		b.Health.Routines = true
		return
	}

	b.Health.Routines = false
	b.Logger.Warn("Routines surpassed threshold!", "threshold", limit, "running", amount)
}
