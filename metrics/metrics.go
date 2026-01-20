// Package METRICS provides Prometheus metrics.
package metrics

import (
	"net/http"
	"regexp"

	"github.com/Shoowa/vamos/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// listOfMetrics returns gauges and counters defined in the library.
func listOfMetrics() []prometheus.Collector {
	return []prometheus.Collector{
		HttpRequestCounter,
		HttpRequestsGauge,
	}
}

// addRuntimeMetrics can't return the internal.GoCollectorOptions, so this
// function returns a struct that can accept the internal structs. When an
// developer decides against using any of the runtime metrics, then the
// NewGoCollector will not be created.
func addRuntimeMetrics(toggles *config.Metrics) prometheus.Collector {
	desiredRules := []collectors.GoRuntimeMetricsRule{}
	if toggles.GarbageCollection {
		desiredRules = append(desiredRules, collectors.MetricsGC)
	}
	if toggles.Memory {
		desiredRules = append(desiredRules, collectors.MetricsMemory)
	}
	if toggles.Scheduler {
		desiredRules = append(desiredRules, collectors.MetricsScheduler)
	}
	if toggles.Cpu {
		cpu := regexp.MustCompile(`^/cpu/classes/.*`)
		cpuRule := collectors.GoRuntimeMetricsRule{Matcher: cpu}
		desiredRules = append(desiredRules, cpuRule)
	}
	if toggles.Lock {
		lock := regexp.MustCompile(`^/sync/mutex/.*`)
		lockRule := collectors.GoRuntimeMetricsRule{Matcher: lock}
		desiredRules = append(desiredRules, lockRule)
	}

	if len(desiredRules) == 0 {
		return nil
	}

	rules := collectors.WithGoCollectorRuntimeMetrics(desiredRules...)
	noOldMemStats := collectors.WithGoCollectorMemStatsMetricsDisabled()

	return collectors.NewGoCollector(noOldMemStats, rules)
}

// createLoadedRegistry reads the config file directly, breaking the convention
// of accepting a config struct from the main function. I chose to do this,
// because the custom registry is a package variable. And it is much easier to
// add metrics to a package variable.
func createLoadedRegistry() *prometheus.Registry {
	toggles := config.Read().Metrics
	runtimeCollector := addRuntimeMetrics(toggles)

	reg := prometheus.NewRegistry()
	if runtimeCollector != nil {
		reg.MustRegister(runtimeCollector)
	}

	if toggles.Process {
		opt := collectors.ProcessCollectorOpts{}
		p := collectors.NewProcessCollector(opt)
		reg.MustRegister(p)
	}

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

// CreateCounter registers a custom counter.
func CreateCounter(name string, help string) prometheus.Counter {
	opts := prometheus.CounterOpts{
		Name: name,
		Help: help,
	}

	counter := prometheus.NewCounter(opts)
	registry.MustRegister(counter)
	return counter
}

// CreateGauge registers a custom gauge.
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

// CreateHistogram registers a custom histogram.
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

// CreateSummary registers a custom summary.
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

// HistogramWithTimer bundles a graph with a convenient timer.
type HistogramWithTimer struct {
	Graph prometheus.Histogram
	Timer func() *prometheus.Timer
}

// CreateHistogramWithTimer registers a histogram and bundles it with a timer
// that can be convenientlay invoked by a developer.
func CreateHistogramWithTimer(ns, ss, name, help string, buckets []float64) *HistogramWithTimer {
	graph := CreateHistogram(ns, ss, name, help, buckets)
	timer := func() *prometheus.Timer { return prometheus.NewTimer(graph) }
	return &HistogramWithTimer{graph, timer}
}
