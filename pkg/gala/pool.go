package gala

import (
	"runtime"

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

// Resize updates worker concurrency when positive
func (p *Pool) Resize(maxWorkers int) {
	if p == nil || maxWorkers <= 0 {
		return
	}

	p.maxWorkers = maxWorkers
	p.pool.Resize(maxWorkers)
}

// WaitForIdle blocks until queued and running tasks drain
func (p *Pool) WaitForIdle() {
	if p == nil {
		return
	}

	for p.pool.RunningWorkers() > 0 || p.pool.WaitingTasks() > 0 {
		runtime.Gosched()
	}
}
