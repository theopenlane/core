package soiree

import (
	"math"
	"sync"

	"github.com/alitto/pond/v2"
)

// Pool is an interface for a worker pool
type Pool interface {
	// Submit submits a task to the worker pool
	Submit(task func())
	// Running returns the number of running workers in the pool
	Running() int64
	// Release stops all workers in the pool and waits for them to finish
	Release()
	// Stop causes this pool to stop accepting new tasks and signals all workers to exit
	Stop()
	// SubmittedTasks returns the number of tasks submitted to the pool
	SubmittedTasks() int
	// WaitingTasks returns the number of tasks waiting in the pool
	WaitingTasks() int
	// SuccessfulTasks returns the number of tasks that completed successfully
	SuccessfulTasks() int
	// FailedTasks returns the number of tasks that completed with a panic
	FailedTasks() int
	// CompletedTasks returns the number of tasks that completed either successfully or with a panic
	CompletedTasks() int
}

// PondPool is a worker pool implementation using the pond library
type PondPool struct {
	// pool is the worker pool
	pool pond.Pool
	// name is the name of the pool used in metrics
	name string
	// MaxWorkers is the maximum number of workers in the pool
	MaxWorkers int `json:"maxWorkers" koanf:"maxWorkers" default:"100"`
	// opts are the options for the pool
	opts []pond.Option
}

// NewPondPool creates a new worker pool using the pond library
func NewPondPool(opts ...PoolOptions) *PondPool {
	p := &PondPool{}

	for _, opt := range opts {
		opt(p)
	}

	p.pool = pond.NewPool(p.MaxWorkers, p.opts...)

	return p
}

// PoolOptions is a type for setting options on the pool
type PoolOptions func(*PondPool)

// WithName sets the name of the pool
func WithName(name string) PoolOptions {
	return func(p *PondPool) {
		p.name = name
	}
}

// WithMaxWorkers sets the maximum number of workers in the pool
func WithMaxWorkers(maxWorkers int) PoolOptions {
	return func(p *PondPool) {
		p.MaxWorkers = maxWorkers
	}
}

// WithOptions sets the options for the pool
func WithOptions(opts ...pond.Option) PoolOptions {
	return func(p *PondPool) {
		p.opts = opts
	}
}

// Submit submits a task to the worker pool
func (p *PondPool) Submit(task func()) {
	p.pool.Submit(task)
}

// SubmitMultipleAndWait submits multiple tasks to the worker pool and waits for all them to finish
func (p *PondPool) SubmitMultipleAndWait(task []func()) {
	wg := new(sync.WaitGroup)

	for _, t := range task {
		wg.Add(1)

		p.pool.Submit(func() {
			// Decrement the counter when the goroutine completes.
			defer wg.Done()
			// Execute the task
			t()
		})
	}

	wg.Wait()
}

// Running returns the number of running workers in the pool
func (p *PondPool) Running() int64 {
	return p.pool.RunningWorkers()
}

// Release stops all workers in the pool and waits for them to finish
func (p *PondPool) Release() {
	p.pool.StopAndWait()
}

// Stop causes this pool to stop accepting new tasks and signals all workers to exit
// Tasks being executed by workers will continue until completion (unless the process is terminated)
// Tasks in the queue will not be executed (so will drop any buffered tasks - ideally use Release)
func (p *PondPool) Stop() {
	p.pool.Stop()
}

// SubmittedTasks returns the number of tasks submitted to the pool
func (p *PondPool) SubmittedTasks() int {
	return uintToInt(p.pool.SubmittedTasks())
}

// WaitingTasks returns the number of tasks waiting in the pool
func (p *PondPool) WaitingTasks() int {
	waitingTasks := p.pool.WaitingTasks()
	maxInt := uint64(math.MaxInt)

	if waitingTasks > maxInt {
		return math.MaxInt
	}

	return int(waitingTasks) // nolint:gosec
}

// SuccessfulTasks returns the number of tasks that completed successfully
func (p *PondPool) SuccessfulTasks() int {
	return uintToInt(p.pool.SuccessfulTasks())
}

// FailedTasks returns the number of tasks that completed with a panic
func (p *PondPool) FailedTasks() int {
	return uintToInt(p.pool.FailedTasks())
}

// CompletedTasks returns the number of tasks that completed either successfully or with a panic
func (p *PondPool) CompletedTasks() int {
	return uintToInt(p.pool.CompletedTasks())
}

func uintToInt(v uint64) int {
	if v > uint64(math.MaxInt) {
		return math.MaxInt
	}

	return int(v) // nolint:gosec
}
