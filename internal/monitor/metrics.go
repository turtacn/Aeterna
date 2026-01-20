package monitor

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/turtacn/Aeterna/pkg/logger"
)

var (
	// HandoverDuration tracks the time taken for a hot relay handover in seconds.
	HandoverDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "aeterna_handover_duration_seconds",
		Help: "Time taken for hot relay handover",
	})
	// RestartTotal tracks the total number of process restarts, partitioned by reason.
	RestartTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "aeterna_restarts_total",
		Help: "Total number of process restarts",
	}, []string{"reason"})
)

// InitMetrics registers Prometheus metrics and starts an HTTP server to expose them.
// It takes an address string (e.g., ":9090") on which to listen for requests.
func InitMetrics(addr string) {
	prometheus.MustRegister(HandoverDuration)
	prometheus.MustRegister(RestartTotal)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		logger.Log.Info("Metrics server starting", "addr", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			logger.Log.Error("Metrics server failed", "err", err)
		}
	}()
}

// Personal.AI order the ending
