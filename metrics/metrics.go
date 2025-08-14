// Package METRICS provides Prometheus metrics.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var readAuthorOpts = prometheus.CounterOpts{
	Name: "read_author_count",
	Help: "amount readAuthor requests",
}

var ReadAuthorCounter = prometheus.NewCounter(readAuthorOpts)

func Register() {
	prometheus.MustRegister(ReadAuthorCounter)
}
