package utils

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	Requests *prometheus.CounterVec
}

func NewMetrics() *Metrics {
	requests := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path"},
	)

	prometheus.MustRegister(requests)

	return &Metrics{
		Requests: requests,
	}
}
