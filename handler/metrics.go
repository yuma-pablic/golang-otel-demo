package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterMetricsRoute(r chi.Router) {
	r.Handle("/metrics", promhttp.Handler())
}
