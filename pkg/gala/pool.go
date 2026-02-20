package gala

import (
	"github.com/alitto/pond/v2"
	"github.com/prometheus/client_golang/prometheus"
)

const defaultPoolWorkers = 200

// Pool is a lightweight in-memory task pool exposed from gala
type Pool struct {
	pool       pond.Pool
	maxWorkers int
	name       string
	metricsReg prometheus.Registerer
}

// PoolOption configures a pool instance
type PoolOption func(*Pool)

// WithWorkers sets maximum worker concurrency
func WithWorkers(n int) PoolOption {
	return func(p *Pool) {
		if n > 0 {
			p.maxWorkers = n
		}
	}
}

// WithPoolName sets the pool name (reserved for metrics labeling)
func WithPoolName(name string) PoolOption {
	return func(p *Pool) {
		p.name = name
	}
}

// WithPoolMetricsRegisterer stores the target registerer for future metrics wiring
func WithPoolMetricsRegisterer(reg prometheus.Registerer) PoolOption {
	return func(p *Pool) {
		p.metricsReg = reg
	}
}

// NewPool creates an in-memory task pool
func NewPool(opts ...PoolOption) *Pool {
	p := &Pool{
		maxWorkers: defaultPoolWorkers,
		metricsReg: prometheus.DefaultRegisterer,
	}

	for _, opt := range opts {
		opt(p)
	}

	p.pool = pond.NewPool(p.maxWorkers)

	if p.metricsReg != nil && p.name != "" {
		collector := newPoolCollector(p.pool, p.name)
		p.metricsReg.MustRegister(collector)
	}

	return p
}

// Submit schedules one task
func (p *Pool) Submit(task func()) {
	if p == nil || task == nil {
		return
	}

	p.pool.Submit(task)
}

// SubmitMultipleAndWait schedules all tasks and waits for completion
func (p *Pool) SubmitMultipleAndWait(tasks []func()) error {
	if p == nil || len(tasks) == 0 {
		return nil
	}

	group := p.pool.NewGroup()
	for _, task := range tasks {
		if task == nil {
			continue
		}

		group.Submit(task)
	}

	return group.Wait()
}

// Release stops workers and waits for completion
func (p *Pool) Release() {
	if p == nil {
		return
	}

	p.pool.StopAndWait()
}

// poolCollector implements prometheus.Collector to expose pool metrics at scrape time
type poolCollector struct {
	pool pond.Pool
	name string

	runningWorkers *prometheus.Desc
	waitingTasks   *prometheus.Desc
	submittedTasks *prometheus.Desc
	completedTasks *prometheus.Desc
	failedTasks    *prometheus.Desc
}

// newPoolCollector creates a collector that reads pool stats on each Prometheus scrape
func newPoolCollector(pool pond.Pool, name string) *poolCollector {
	labels := prometheus.Labels{"pool": name}

	return &poolCollector{
		pool:           pool,
		name:           name,
		runningWorkers: prometheus.NewDesc("gala_pool_running_workers", "Number of workers currently executing tasks", nil, labels),
		waitingTasks:   prometheus.NewDesc("gala_pool_waiting_tasks", "Number of tasks waiting in the queue", nil, labels),
		submittedTasks: prometheus.NewDesc("gala_pool_submitted_tasks_total", "Total number of tasks submitted to the pool", nil, labels),
		completedTasks: prometheus.NewDesc("gala_pool_completed_tasks_total", "Total number of tasks completed by the pool", nil, labels),
		failedTasks:    prometheus.NewDesc("gala_pool_failed_tasks_total", "Total number of tasks that failed", nil, labels),
	}
}

// Describe sends descriptor metadata to the channel
func (c *poolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.runningWorkers
	ch <- c.waitingTasks
	ch <- c.submittedTasks
	ch <- c.completedTasks
	ch <- c.failedTasks
}

// Collect reads current pool stats and sends metrics to the channel
func (c *poolCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(c.runningWorkers, prometheus.GaugeValue, float64(c.pool.RunningWorkers()))
	ch <- prometheus.MustNewConstMetric(c.waitingTasks, prometheus.GaugeValue, float64(c.pool.WaitingTasks()))
	ch <- prometheus.MustNewConstMetric(c.submittedTasks, prometheus.CounterValue, float64(c.pool.SubmittedTasks()))
	ch <- prometheus.MustNewConstMetric(c.completedTasks, prometheus.CounterValue, float64(c.pool.CompletedTasks()))
	ch <- prometheus.MustNewConstMetric(c.failedTasks, prometheus.CounterValue, float64(c.pool.FailedTasks()))
}
