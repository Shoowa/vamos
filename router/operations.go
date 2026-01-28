package router

import (
	"context"
	"log/slog"
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
func (b *Backbone) PingDB(health *Health) {
	timer, cancel := context.WithTimeout(context.Background(), TIMEOUT_PING)
	defer cancel()

	err := b.DbHandle.Ping(timer)
	if err != nil {
		health.Rdbms = false
		b.Logger.Error("Failed ping", "Rdbms", err.Error())
		return
	}
	health.Rdbms = true
}

// SetupHealthChecks reads configured values, and leverages closures to apply
// them to Backbone methods designed to run periodically. Each of those methods
// evaluates one condition in the application or a dependency.
func setupHealthChecks(cfg *config.Config, b *Backbone) *Health {
	// Create the health record.
	health := new(Health)
	health.Rdbms = false
	health.Heap = true
	health.Routines = true

	// Report status of connection upon ignition.
	b.PingDB(health)

	pingDbTimer := time.Duration(cfg.Health.PingDbTimer)
	heapTimer := time.Duration(cfg.Health.HeapTimer)
	routinesTimer := time.Duration(cfg.Health.RoutTimer)

	// Use closure to configure method CheckHeapSize.
	heapSize := 1024 * 1024 * cfg.Health.HeapSize
	checkHeapSize := func() { b.checkHeapSize(health, heapSize) }

	// Use closure to configure CheckNumRoutines.
	limit := runtime.NumCPU() * cfg.Health.RoutinesPerCore
	checkNumRoutines := func() { checkNumRoutines(health, limit, b.Logger) }

	// Use  closure to add the Health Record to the pinger.
	pingDB := func() { b.PingDB(health) }

	go beep(pingDbTimer, pingDB)
	go beep(heapTimer, checkHeapSize)
	go beep(routinesTimer, checkNumRoutines)

	return health
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
func (b *Backbone) checkHeapSize(health *Health, threshold uint64) {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	if stats.HeapAlloc < threshold {
		health.Heap = true
		return
	}

	health.Heap = false
	b.Logger.Warn("Heap surpassed threshold!", "threshold", threshold, "allocated", stats.HeapAlloc)
	err := pprof.WriteHeapProfile(b)
	if err != nil {
		b.Logger.Error("Error writing heap profile", "ERR:", err.Error())
	}
}

// CheckNumRoutines reads the runtime to determine whether or not the amount of
// routines surpasses a configured value, then fail a health check.
func checkNumRoutines(health *Health, limit int, logger *slog.Logger) {
	amount := runtime.NumGoroutine()
	if amount < limit {
		health.Routines = true
		return
	}

	health.Routines = false
	logger.Warn("Routines surpassed threshold!", "threshold", limit, "running", amount)
}
