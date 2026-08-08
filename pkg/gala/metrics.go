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

// initListenerDeliveries creates a collector for total listener deliveries
func initListenerDeliveries() *prometheus.CounterVec {
	return promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gala_listener_deliveries_total",
			Help: "Number of gala listener deliveries",
		},
		listenerMetricLabels,
	)
}

// initListenerFailures creates a collector for failed listener deliveries
func initListenerFailures() *prometheus.CounterVec {
	return promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gala_listener_failures_total",
			Help: "Number of failed gala listener deliveries",
		},
		listenerMetricLabels,
	)
}

// initListenerDuration creates a collector for listener execution duration
func initListenerDuration() *prometheus.HistogramVec {
	return promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "gala_listener_duration_seconds",
			Help: "Time in seconds per gala listener delivery",
		},
		listenerMetricLabels,
	)
}

// initialize the collectors, prometheus will register them automatically
var (
	listenerDeliveries = initListenerDeliveries()
	listenerFailures   = initListenerFailures()
	listenerDuration   = initListenerDuration()
)
