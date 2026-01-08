package soiree

import (
	"context"
	"errors"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// redisMetrics holds prometheus collectors for redis store operations
type redisMetrics struct {
	redisEventsPersisted  prometheus.Counter
	redisEventsDequeued   prometheus.Counter
	redisResultsPersisted prometheus.Counter
	redisQueueLength      prometheus.Gauge
}

// poolMetrics holds prometheus collectors for worker pool activity
type poolMetrics struct {
	tasksSubmitted prometheus.Counter
	tasksStarted   prometheus.Counter
	tasksCompleted prometheus.Counter
	tasksPanicked  prometheus.Counter
	tasksQueued    prometheus.Gauge
	tasksRunning   prometheus.Gauge
}

const (
	poolMetricsLabelName = "pool"
	defaultPoolName      = "default"
)

// poolMetricsVec holds labeled prometheus collectors for worker pool activity
type poolMetricsVec struct {
	tasksSubmitted *prometheus.CounterVec
	tasksStarted   *prometheus.CounterVec
	tasksCompleted *prometheus.CounterVec
	tasksPanicked  *prometheus.CounterVec
	tasksQueued    *prometheus.GaugeVec
	tasksRunning   *prometheus.GaugeVec
}

// newRedisMetrics creates a new set of redis metrics registered with the given registerer
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

// newPoolMetricsVec creates labeled pool metrics registered with the given registerer.
func newPoolMetricsVec(reg prometheus.Registerer) *poolMetricsVec {
	labelNames := []string{poolMetricsLabelName}
	return &poolMetricsVec{
		tasksSubmitted: registerCounterVec(reg, prometheus.CounterOpts{
			Name: "soiree_pool_tasks_submitted_total",
			Help: "Total number of tasks submitted to the soiree worker pool",
		}, labelNames),
		tasksStarted: registerCounterVec(reg, prometheus.CounterOpts{
			Name: "soiree_pool_tasks_started_total",
			Help: "Total number of tasks started by the soiree worker pool",
		}, labelNames),
		tasksCompleted: registerCounterVec(reg, prometheus.CounterOpts{
			Name: "soiree_pool_tasks_completed_total",
			Help: "Total number of tasks completed by the soiree worker pool",
		}, labelNames),
		tasksPanicked: registerCounterVec(reg, prometheus.CounterOpts{
			Name: "soiree_pool_tasks_panicked_total",
			Help: "Total number of tasks that panicked in the soiree worker pool",
		}, labelNames),
		tasksQueued: registerGaugeVec(reg, prometheus.GaugeOpts{
			Name: "soiree_pool_tasks_queued",
			Help: "Current number of tasks waiting to start in the soiree worker pool",
		}, labelNames),
		tasksRunning: registerGaugeVec(reg, prometheus.GaugeOpts{
			Name: "soiree_pool_tasks_running",
			Help: "Current number of tasks running in the soiree worker pool",
		}, labelNames),
	}
}

// Default metrics for production use
var defaultRedisMetrics = newRedisMetrics(prometheus.DefaultRegisterer)

// registerCounterVec creates and registers a CounterVec with the given registerer, returning an existing collector if already registered
func registerCounterVec(reg prometheus.Registerer, opts prometheus.CounterOpts, labels []string) *prometheus.CounterVec {
	vec := prometheus.NewCounterVec(opts, labels)
	if err := reg.Register(vec); err != nil {
		var already prometheus.AlreadyRegisteredError
		if errors.As(err, &already) {
			if existing, ok := already.ExistingCollector.(*prometheus.CounterVec); ok {
				return existing
			}
		}
		panic(err)
	}
	return vec
}

// registerGaugeVec creates and registers a GaugeVec with the given registerer, returning an existing collector if already registered
func registerGaugeVec(reg prometheus.Registerer, opts prometheus.GaugeOpts, labels []string) *prometheus.GaugeVec {
	vec := prometheus.NewGaugeVec(opts, labels)
	if err := reg.Register(vec); err != nil {
		var already prometheus.AlreadyRegisteredError
		if errors.As(err, &already) {
			if existing, ok := already.ExistingCollector.(*prometheus.GaugeVec); ok {
				return existing
			}
		}
		panic(err)
	}
	return vec
}

// poolMetricsFor returns pool metrics for the given pool name, creating them if necessary
func poolMetricsFor(reg prometheus.Registerer, poolName string) *poolMetrics {
	if reg == nil {
		return nil
	}
	if poolName == "" {
		poolName = defaultPoolName
	}

	metricsVec := newPoolMetricsVec(reg)
	return &poolMetrics{
		tasksSubmitted: metricsVec.tasksSubmitted.WithLabelValues(poolName),
		tasksStarted:   metricsVec.tasksStarted.WithLabelValues(poolName),
		tasksCompleted: metricsVec.tasksCompleted.WithLabelValues(poolName),
		tasksPanicked:  metricsVec.tasksPanicked.WithLabelValues(poolName),
		tasksQueued:    metricsVec.tasksQueued.WithLabelValues(poolName),
		tasksRunning:   metricsVec.tasksRunning.WithLabelValues(poolName),
	}
}

// initQueueLength initializes the queue length metric from the current redis state
func (s *RedisStore) initQueueLength() {
	n, err := s.client.LLen(context.Background(), "soiree:queue").Result()
	if err == nil {
		s.metrics.redisQueueLength.Set(float64(n))
	}
}
