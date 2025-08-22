// Package METRICS provides Prometheus metrics.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

func requestCounter() *prometheus.CounterVec {
	options := prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Amount of HTTP requests received.",
	}
	labels := []string{"status", "path", "method"}
	counter := prometheus.NewCounterVec(options, labels)
	prometheus.MustRegister(counter)
	return counter
}

var HttpRequestCounter = requestCounter()

func CreateCounter(name string, help string) prometheus.Counter {
	opts := prometheus.CounterOpts{
		Name: name,
		Help: help,
	}

	counter := prometheus.NewCounter(opts)
	prometheus.MustRegister(counter)
	return counter
}
