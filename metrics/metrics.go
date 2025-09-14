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

func connectionsGauge() prometheus.Gauge {
	options := prometheus.GaugeOpts{
		Name: "http_active_requests",
		Help: "Amount of active HTTP requests.",
	}
	gauge := prometheus.NewGauge(options)
	prometheus.MustRegister(gauge)
	return gauge
}

var HttpRequestsGauge = connectionsGauge()

func CreateCounter(name string, help string) prometheus.Counter {
	opts := prometheus.CounterOpts{
		Name: name,
		Help: help,
	}

	counter := prometheus.NewCounter(opts)
	prometheus.MustRegister(counter)
	return counter
}

func CreateGauge(ns, ss, name, help string) prometheus.Gauge {
	opts := prometheus.GaugeOpts{
		Namespace: ns,
		Subsystem: ss,
		Name:      name,
		Help:      help,
	}

	gauge := prometheus.NewGauge(opts)
	prometheus.MustRegister(gauge)
	return gauge
}
