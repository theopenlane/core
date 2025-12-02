package soiree

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// NewStatsCollector registers Prometheus metrics for the PondPool
func (p *PondPool) NewStatsCollector() {
	// add a default name if none is provided
	name := p.name
	if name == "" {
		name = "default"
	}

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: fmt.Sprintf("pool_%s_workers_running", name),
			Help: "Number of running worker goroutines",
		},
		func() float64 {
			return float64(p.Running())
		}))

	// Task metrics
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: fmt.Sprintf("pool_%s_tasks_submitted_total", name),
			Help: "Number of tasks submitted",
		},
		func() float64 {
			return float64(p.SubmittedTasks())
		}))
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: fmt.Sprintf("pool_%s_tasks_waiting_total", name),
			Help: "Number of tasks waiting in the queue",
		},
		func() float64 {
			return float64(p.WaitingTasks())
		}))
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: fmt.Sprintf("pool_%s_tasks_successful_total", name),
			Help: "Number of tasks that completed successfully",
		},
		func() float64 {
			return float64(p.SuccessfulTasks())
		}))
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: fmt.Sprintf("pool_%s_tasks_failed_total", name),
			Help: "Number of tasks that completed with panic",
		},
		func() float64 {
			return float64(p.FailedTasks())
		}))
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: fmt.Sprintf("pool_%s_tasks_completed_total", name),
			Help: "Number of tasks that completed either successfully or with panic",
		},
		func() float64 {
			return float64(p.CompletedTasks())
		}))
}

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
