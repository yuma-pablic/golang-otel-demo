package utils

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	Requests *prometheus.CounterVec
	Duration *prometheus.HistogramVec
}

func NewMetrics() *Metrics {
	requests := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path"},
	)

	duration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 10), // 10バケツ
		},
		[]string{"path"},
	)

	prometheus.MustRegister(requests, duration)

	return &Metrics{
		Requests: requests,
		Duration: duration,
	}
}
