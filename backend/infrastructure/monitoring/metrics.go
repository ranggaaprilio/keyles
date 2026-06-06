package monitoring

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	registry = prometheus.NewRegistry()

	securityEventsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "keyles_security_events_total",
			Help: "Total number of security events",
		},
		[]string{"event_type"},
	)

	securityEventDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "keyles_security_event_duration_seconds",
			Help:    "Duration of security event handling in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"event_type"},
	)
)

func init() {
	registry.MustRegister(securityEventsTotal)
	registry.MustRegister(securityEventDuration)
}

// GetMetricsHandler returns an HTTP handler for the /metrics endpoint
func GetMetricsHandler() http.Handler {
	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}

// IncrementSecurityEvent increments the security events counter for the given event type
func IncrementSecurityEvent(eventType string) {
	securityEventsTotal.WithLabelValues(eventType).Inc()
}

// ObserveSecurityEventDuration records the duration of a security event
func ObserveSecurityEventDuration(eventType string, duration time.Duration) {
	securityEventDuration.WithLabelValues(eventType).Observe(duration.Seconds())
}
