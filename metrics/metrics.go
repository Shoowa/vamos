// Package METRICS provides Prometheus metrics.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

func CreateCounter(name string, help string) prometheus.Counter {
	opts := prometheus.CounterOpts{
		Name: name,
		Help: help,
	}

	counter := prometheus.NewCounter(opts)
	prometheus.MustRegister(counter)
	return counter
}
