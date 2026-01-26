package soiree

import (
	"runtime"

	"github.com/alitto/pond/v2"
	"github.com/prometheus/client_golang/prometheus"
)

// defaultPoolWorkers is the default number of workers in a pool
const defaultPoolWorkers = 10

// Pool is a worker pool implementation using the pond library
type Pool struct {
	pool       pond.Pool
	maxWorkers int
	name       string
	metricsReg prometheus.Registerer
	metrics    *poolMetrics
}

// PoolOption configures a Pool
type PoolOption func(*Pool)

// WithWorkers sets the maximum number of workers in the pool
func WithWorkers(n int) PoolOption {
	return func(p *Pool) {
		p.maxWorkers = n
	}
}

// WithPoolMetricsRegisterer configures pool metrics using the provided registerer (nil disables metrics)
func WithPoolMetricsRegisterer(reg prometheus.Registerer) PoolOption {
	return func(p *Pool) {
		p.metricsReg = reg
	}
}

// WithPoolName sets the pool name used for metrics labeling
func WithPoolName(name string) PoolOption {
	return func(p *Pool) {
		p.name = name
	}
}

// NewPool creates a new worker pool with the given options
func NewPool(opts ...PoolOption) *Pool {
	p := &Pool{
		maxWorkers: defaultPoolWorkers,
		metricsReg: prometheus.DefaultRegisterer,
	}

	for _, opt := range opts {
		opt(p)
	}

	p.metrics = poolMetricsFor(p.metricsReg, p.name)
	p.pool = pond.NewPool(p.maxWorkers)

	return p
}

// Submit submits a task to the worker pool
func (p *Pool) Submit(task func()) {
	p.trackSubmission()
	p.pool.Submit(p.wrapTask(task))
}

// SubmitMultipleAndWait submits multiple tasks and waits for all to complete
func (p *Pool) SubmitMultipleAndWait(tasks []func()) error {
	group := p.pool.NewGroup()
	for _, task := range tasks {
		p.trackSubmission()
		group.Submit(p.wrapTask(task))
	}

	return group.Wait()
}

// Release stops all workers in the pool and waits for them to finish
func (p *Pool) Release() {
	p.pool.StopAndWait()
}

// Resize adjusts the maximum number of workers in the pool
func (p *Pool) Resize(maxWorkers int) {
	if maxWorkers <= 0 {
		return
	}

	p.maxWorkers = maxWorkers
	p.pool.Resize(maxWorkers)
}

// WaitForIdle blocks until all submitted tasks have completed
func (p *Pool) WaitForIdle() {
	for p.pool.RunningWorkers() > 0 || p.pool.WaitingTasks() > 0 {
		// yield to allow workers to make progress
		runtime.Gosched()
	}
}

// trackSubmission updates metrics when a task is submitted
func (p *Pool) trackSubmission() {
	if p.metrics == nil {
		return
	}

	p.metrics.tasksSubmitted.Inc()
	p.metrics.tasksQueued.Inc()
}

// wrapTask wraps a task function to track metrics around its execution
func (p *Pool) wrapTask(task func()) func() {
	if p.metrics == nil {
		return task
	}

	return func() {
		p.metrics.tasksQueued.Dec()
		p.metrics.tasksRunning.Inc()
		p.metrics.tasksStarted.Inc()

		defer func() {
			p.metrics.tasksRunning.Dec()
			if r := recover(); r != nil {
				p.metrics.tasksPanicked.Inc()
				panic(r)
			}
			p.metrics.tasksCompleted.Inc()
		}()

		task()
	}
}
