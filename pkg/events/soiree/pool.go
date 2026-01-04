package soiree

import (
	"github.com/alitto/pond/v2"
)

// pool is an interface for a worker pool
type pool interface {
	Submit(task func())
	Release()
}

// PondPool is a worker pool implementation using the pond library
type PondPool struct {
	pool       pond.Pool
	maxWorkers int
}

// PondPoolOption configures a PondPool
type PondPoolOption func(*PondPool)

// WithMaxWorkers sets the maximum number of workers in the pool
func WithMaxWorkers(n int) PondPoolOption {
	return func(p *PondPool) {
		p.maxWorkers = n
	}
}

// WithName sets a name for the pool (currently unused, kept for API compatibility)
func WithName(_ string) PondPoolOption {
	return func(_ *PondPool) {}
}

// WithOptions applies multiple options at once
func WithOptions(opts ...PondPoolOption) PondPoolOption {
	return func(p *PondPool) {
		for _, opt := range opts {
			opt(p)
		}
	}
}

// NewPondPool creates a new worker pool with the given options
func NewPondPool(opts ...PondPoolOption) *PondPool {
	p := &PondPool{maxWorkers: 10}
	for _, opt := range opts {
		opt(p)
	}
	p.pool = pond.NewPool(p.maxWorkers)
	return p
}

// newPondPool creates a new worker pool with the given number of workers (internal use)
func newPondPool(maxWorkers int) *PondPool {
	return NewPondPool(WithMaxWorkers(maxWorkers))
}

// Submit submits a task to the worker pool
func (p *PondPool) Submit(task func()) {
	p.pool.Submit(task)
}

// SubmitMultipleAndWait submits multiple tasks and waits for all to complete
func (p *PondPool) SubmitMultipleAndWait(tasks []func()) {
	group := p.pool.NewGroup()
	for _, task := range tasks {
		group.Submit(task)
	}
	group.Wait()
}

// Release stops all workers in the pool and waits for them to finish
func (p *PondPool) Release() {
	p.pool.StopAndWait()
}

// NewStatsCollector is a no-op for API compatibility
func (p *PondPool) NewStatsCollector() {}
