package usage

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/theopenlane/core/pkg/enums"
)

var (
	usageUpdateCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "usage_update_total",
		Help: "Number of usage value updates",
	}, []string{"type", "op"})

	usageLimitCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "usage_limit_changes_total",
		Help: "Number of usage limit modifications",
	}, []string{"type", "op"})
)

// RecordUsageUpdate increments the usage update counter for the given type and operation
func RecordUsageUpdate(t enums.UsageType, op string) {
	usageUpdateCounter.WithLabelValues(t.String(), op).Inc()
}

// RecordLimitChange increments the usage limit change counter for the given type and operation
func RecordLimitChange(t enums.UsageType, op string) {
	usageLimitCounter.WithLabelValues(t.String(), op).Inc()
}
