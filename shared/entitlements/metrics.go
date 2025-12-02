package entitlements

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	stripeRequestCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "stripe_requests_total",
		Help: "Total number of Stripe API requests",
	}, []string{"type", "status"})

	stripeRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "stripe_request_duration_seconds",
		Help:    "Duration of Stripe API requests",
		Buckets: prometheus.DefBuckets,
	}, []string{"type", "status"})
)
