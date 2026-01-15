package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RequestCounter counts total requests by status (success/blocked) and source (input/output)
	RequestCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "teachme_requests_total",
		Help: "The total number of requests processed",
	}, []string{"status", "source"})

	// LatencyHistogram tracks how long requests take
	LatencyHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "teachme_request_duration_seconds",
		Help:    "Time taken to process request",
		Buckets: prometheus.DefBuckets,
	}, []string{"status"})
)
