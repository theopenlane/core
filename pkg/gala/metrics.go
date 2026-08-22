package gala

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	// metricLabelTopic is the metrics label carrying the topic name
	metricLabelTopic = "topic"
	// metricLabelListener is the metrics label carrying the listener definition name
	metricLabelListener = "listener"
	// metricLabelOperation is the metrics label carrying the payload operation
	metricLabelOperation = "operation"
)

// listenerMetricLabels are the labels applied to every gala listener collector
var listenerMetricLabels = []string{metricLabelTopic, metricLabelListener, metricLabelOperation}

// listener delivery collectors, registered automatically by promauto
var (
	listenerDeliveries = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gala_listener_deliveries_total",
			Help: "Number of gala listener deliveries",
		},
		listenerMetricLabels,
	)
	listenerFailures = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gala_listener_failures_total",
			Help: "Number of failed gala listener deliveries",
		},
		listenerMetricLabels,
	)
	listenerDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "gala_listener_duration_seconds",
			Help: "Time in seconds per gala listener delivery",
		},
		listenerMetricLabels,
	)
)
