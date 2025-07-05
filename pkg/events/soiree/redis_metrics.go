package soiree

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type redisMetrics struct {
	redisEventsPersisted  prometheus.Counter
	redisEventsDequeued   prometheus.Counter
	redisResultsPersisted prometheus.Counter
	redisQueueLength      prometheus.Gauge
}

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

// Default metrics for production use
var defaultRedisMetrics = newRedisMetrics(prometheus.DefaultRegisterer)

func (s *RedisStore) initQueueLength() {
	n, err := s.client.LLen(context.Background(), "soiree:queue").Result()
	if err == nil {
		s.metrics.redisQueueLength.Set(float64(n))
	}
}
