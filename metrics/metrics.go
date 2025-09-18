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

func CreateHistogram(ns, ss, name, help string, buckets []float64) prometheus.Histogram {
	opts := prometheus.HistogramOpts{
		Namespace: ns,
		Subsystem: ss,
		Name:      name,
		Help:      help,
		Buckets:   buckets,
	}

	histogram := prometheus.NewHistogram(opts)
	prometheus.MustRegister(histogram)
	return histogram
}

func CreateSummary(ns, ss, name, help string, obj map[float64]float64) prometheus.Summary {
	opts := prometheus.SummaryOpts{
		Namespace:  ns,
		Subsystem:  ss,
		Name:       name,
		Help:       help,
		Objectives: obj,
	}

	sum := prometheus.NewSummary(opts)
	prometheus.MustRegister(sum)
	return sum
}

type HistogramWithTimer struct {
	Graph prometheus.Histogram
	Timer func() *prometheus.Timer
}

func CreateHistogramWithTimer(ns, ss, name, help string, buckets []float64) *HistogramWithTimer {
	graph := CreateHistogram(ns, ss, name, help, buckets)
	timer := func() *prometheus.Timer { return prometheus.NewTimer(graph) }
	return &HistogramWithTimer{graph, timer}
}
