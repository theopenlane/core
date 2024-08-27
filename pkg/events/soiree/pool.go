package soiree

import (
	"sync"
	"time"

	"github.com/alitto/pond"
)

// Pool is an interface for a worker pool
type Pool interface {
	// Submit submits a task to the worker pool
	Submit(task func())
	// Running returns the number of running workers in the pool
	Running() int
	// Release stops all workers in the pool and waits for them to finish
	Release()
	// ReleaseWithDeadline stops this pool and waits until either all tasks in the queue are completed or the given deadline is reached
	ReleaseWithDeadline(deadline time.Duration)
	// Stop causes this pool to stop accepting new tasks and signals all workers to exit
	Stop()
	// IdleWorkers returns the number of idle workers in the pool
	IdleWorkers() int
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
	// StopAndWaitFor stops this pool and waits until either all tasks in the queue are completed or the given deadline is reached
	StopAndWaitFor(deadline time.Duration)
	// SubmitAndWait submits a task to the worker pool and waits for it to finish
	SubmitAndWait(task func())
	// SubmitBefore submits a task to the worker pool before a specified task
	SubmitBefore(task func(), deadline time.Duration)
}

// PondPool is a worker pool implementation using the pond library
type PondPool struct {
	// pool is the worker pool
	pool *pond.WorkerPool
	// name is the name of the pool used in metrics
	name string
}

// NewPondPool creates a new instance of PondPool with the passed options
func NewPondPool(maxWorkers, maxCapacity int, options ...pond.Option) *PondPool {
	return &PondPool{
		pool: pond.New(maxWorkers, maxCapacity, options...),
	}
}

// NewNamedPondPool creates a new instance of PondPool with the passed options and name
func NewNamedPondPool(maxWorkers, maxCapacity int, name string, options ...pond.Option) *PondPool {
	return &PondPool{
		pool: pond.New(maxWorkers, maxCapacity, options...),
		name: name,
	}
}

// Submit submits a task to the worker pool
func (p *PondPool) Submit(task func()) {
	p.pool.Submit(task)
}

// SubmitAndWait submits a task to the worker pool and waits for it to finish
func (p *PondPool) SubmitAndWait(task func()) {
	p.pool.SubmitAndWait(task)
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

// SubmitBefore submits a task to the worker pool before a specified task
func (p *PondPool) SubmitBefore(task func(), deadline time.Duration) {
	p.pool.SubmitBefore(task, deadline)
}

// Running returns the number of running workers in the pool
func (p *PondPool) Running() int {
	return p.pool.RunningWorkers()
}

// Release stops all workers in the pool and waits for them to finish
func (p *PondPool) Release() {
	p.pool.StopAndWait()
}

// ReleaseWithDeadline stops this pool and waits until either all tasks in the queue are completed
// or the given deadline is reached, whichever comes first
func (p *PondPool) ReleaseWithDeadline(deadline time.Duration) {
	p.pool.StopAndWaitFor(deadline)
}

// Stop causes this pool to stop accepting new tasks and signals all workers to exit
// Tasks being executed by workers will continue until completion (unless the process is terminated)
// Tasks in the queue will not be executed (so will drop any buffered tasks - ideally use Release or ReleaseWithDeadline)
func (p *PondPool) Stop() {
	p.pool.Stop()
}

// IdleWorkers returns the number of idle workers in the pool
func (p *PondPool) IdleWorkers() int {
	return p.pool.IdleWorkers()
}

// SubmittedTasks returns the number of tasks submitted to the pool
func (p *PondPool) SubmittedTasks() int {
	return int(p.pool.SubmittedTasks())
}

// WaitingTasks returns the number of tasks waiting in the pool
func (p *PondPool) WaitingTasks() int {
	return int(p.pool.WaitingTasks())
}

// SuccessfulTasks returns the number of tasks that completed successfully
func (p *PondPool) SuccessfulTasks() int {
	return int(p.pool.SuccessfulTasks())
}

// FailedTasks returns the number of tasks that completed with a panic
func (p *PondPool) FailedTasks() int {
	return int(p.pool.FailedTasks())
}

// CompletedTasks returns the number of tasks that completed either successfully or with a panic
func (p *PondPool) CompletedTasks() int {
	return int(p.pool.CompletedTasks())
}

// StopAndWaitFor stops this pool and waits until either all tasks in the queue are completed
// or the given deadline is reached, whichever comes first
func (p *PondPool) StopAndWaitFor(deadline time.Duration) {
	p.pool.StopAndWaitFor(deadline)
}
