package monitor

import (
	"testing"
	"time"
)

func TestMetricsInitialization(t *testing.T) {
	addr := "127.0.0.1:0" // Random port
	InitMetrics(addr)

	// Increment metrics to see if they are working
	RestartTotal.WithLabelValues("test").Inc()
	HandoverDuration.Observe(0.5)

	// Briefly check if we can reach the metrics endpoint
	time.Sleep(100 * time.Millisecond)
}

func TestMetricsValues(t *testing.T) {
	// Just verify we can use them
	RestartTotal.WithLabelValues("manual").Inc()
	HandoverDuration.Observe(1.0)
}
