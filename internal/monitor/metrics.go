package monitor

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/turtacn/Aeterna/pkg/logger"
)

var (
	HandoverDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "aeterna_handover_duration_seconds",
		Help: "Time taken for hot relay handover",
	})
	RestartTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "aeterna_restarts_total",
		Help: "Total number of process restarts",
	}, []string{"reason"})
)

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
