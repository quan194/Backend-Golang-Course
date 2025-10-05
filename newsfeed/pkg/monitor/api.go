package monitor

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	apiStatusCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "newsfeed",
			Subsystem: "api",
			Name:      "status_count",
		},
		[]string{"api", "method", "status"},
	)

	apiStatusLatency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  "newsfeed",
			Subsystem:  "api",
			Name:       "status_latency",
			Objectives: map[float64]float64{0.5: 0.5, 0.9: 0.1, 0.99: 0.001}, // p50 p90 p99
		},
		[]string{"api", "method", "status"},
	)
)

func init() {
	prometheus.MustRegister(
		apiStatusCounter,
		apiStatusLatency,
	)
}

func ExportApiStatus(api, method string, httpCode int) {
	apiStatusCounter.WithLabelValues(api, method, strconv.Itoa(httpCode)).Inc()
}

func ExportApiStatusLatency(api, method string, httpCode int, latency time.Duration) {
	ms := latency.Milliseconds()
	apiStatusLatency.WithLabelValues(api, method, strconv.Itoa(httpCode)).Observe(float64(ms))
}
