package soiree

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// redisMetrics holds Prometheus metrics for Redis-backed Soiree stores
type redisMetrics struct {
	redisEventsPersisted  prometheus.Counter
	redisEventsDequeued   prometheus.Counter
	redisResultsPersisted prometheus.Counter
	redisQueueLength      prometheus.Gauge
}

// newRedisMetrics constructs a new redisMetrics instance with the provided Prometheus registerer
func newRedisMetrics(reg prometheus.Registerer) *redisMetrics {
	return &redisMetrics{
		redisEventsPersisted: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "soiree_redis_events_persisted_total",
			Help: "Total number of events persisted to redis",
		}),
		redisEventsDequeued: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "soiree_redis_events_dequeued_total",
			Help: "Total number of events dequeued from redis",
		}),
		redisResultsPersisted: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "soiree_redis_results_persisted_total",
			Help: "Total number of handler results persisted to redis",
		}),
		redisQueueLength: promauto.With(reg).NewGauge(prometheus.GaugeOpts{
			Name: "soiree_redis_queue_length",
			Help: "Current number of events waiting in redis",
		}),
	}
}

// defaultRedisMetrics holds the default metrics instance used by Redis stores
var defaultRedisMetrics = newRedisMetrics(prometheus.DefaultRegisterer)

// initQueueLength initializes the redisQueueLength gauge with the current length of the Redis queue
func (s *RedisStore) initQueueLength() {
	n, err := s.client.LLen(context.Background(), "soiree:queue").Result()
	if err == nil {
		s.metrics.redisQueueLength.Set(float64(n))
	}
}
