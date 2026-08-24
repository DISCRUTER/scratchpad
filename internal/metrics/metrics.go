package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	RequestCount    *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	ActiveRequests  prometheus.Gauge
}

// Register metrics
func registerMetrics(reg prometheus.Registerer) *Metrics {
	// Set up metrics to be tracked
	m := &Metrics{
		RequestCount: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "scratchpad",
			Name:      "http_request_total",
			Help:      "Total number of http request.",
		}, []string{"method", "endpoint", "status_code"}),
		RequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "scratchpad",
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds",
		}, []string{"method", "endpoint"}),
		ActiveRequests: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "scratchpad",
			Name:      "http_active_requests",
			Help:      "Number of active HTTP requests",
		}),
	}
	// Register the metrics
	reg.MustRegister(m.RequestCount, m.RequestDuration, m.ActiveRequests)
	return m
}

// Get metrics and handler for dependency injection
func GetMetrics() (*Metrics, *http.ServeMux) {
	// Init a new prometheus registerer
	reg := prometheus.NewRegistry()
	// Register Metrics
	m := registerMetrics(reg)
	// Create prometheus servermux
	pMux := http.NewServeMux()
	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg})
	pMux.Handle("GET /metrics", promHandler)

	return m, pMux
}
