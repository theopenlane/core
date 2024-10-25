package soiree

import (
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
}

// NewPondPool creates a new instance of PondPool with the passed options
func NewPondPool(maxWorkers int, options ...pond.Option) *PondPool {
	return &PondPool{
		pool: pond.NewPool(maxWorkers, options...),
	}
}

// NewNamedPondPool creates a new instance of PondPool with the passed options and name
func NewNamedPondPool(maxWorkers int, name string, options ...pond.Option) *PondPool {
	return &PondPool{
		pool: pond.NewPool(maxWorkers, options...),
		name: name,
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
	return int(p.pool.SubmittedTasks()) // nolint:gosec
}

// WaitingTasks returns the number of tasks waiting in the pool
func (p *PondPool) WaitingTasks() int {
	return int(p.pool.WaitingTasks()) // nolint:gosec
}

// SuccessfulTasks returns the number of tasks that completed successfully
func (p *PondPool) SuccessfulTasks() int {
	return int(p.pool.SuccessfulTasks()) // nolint:gosec
}

// FailedTasks returns the number of tasks that completed with a panic
func (p *PondPool) FailedTasks() int {
	return int(p.pool.FailedTasks()) // nolint:gosec
}

// CompletedTasks returns the number of tasks that completed either successfully or with a panic
func (p *PondPool) CompletedTasks() int {
	return int(p.pool.CompletedTasks()) // nolint:gosec
}
