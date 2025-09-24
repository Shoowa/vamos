// Package METRICS provides Prometheus metrics.
package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func listOfMetrics() []prometheus.Collector {
	return []prometheus.Collector{
		HttpRequestCounter,
		HttpRequestsGauge,
	}
}

func createLoadedRegistry() *prometheus.Registry {
	reg := prometheus.NewRegistry()
	metrics := listOfMetrics()
	for _, m := range metrics {
		reg.MustRegister(m)
	}
	return reg
}

var registry = createLoadedRegistry()

func CreateHandler() http.Handler {
	options := promhttp.HandlerOpts{}
	metricsHandler := promhttp.HandlerFor(registry, options)
	return metricsHandler
}

func requestCounter() *prometheus.CounterVec {
	options := prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Amount of HTTP requests received.",
	}
	labels := []string{"status", "path", "method"}
	counter := prometheus.NewCounterVec(options, labels)
	return counter
}

var HttpRequestCounter = requestCounter()

func connectionsGauge() prometheus.Gauge {
	options := prometheus.GaugeOpts{
		Name: "http_active_requests",
		Help: "Amount of active HTTP requests.",
	}
	gauge := prometheus.NewGauge(options)
	return gauge
}

var HttpRequestsGauge = connectionsGauge()

func CreateCounter(name string, help string) prometheus.Counter {
	opts := prometheus.CounterOpts{
		Name: name,
		Help: help,
	}

	counter := prometheus.NewCounter(opts)
	registry.MustRegister(counter)
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
	registry.MustRegister(gauge)
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
	registry.MustRegister(histogram)
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
	registry.MustRegister(sum)
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
