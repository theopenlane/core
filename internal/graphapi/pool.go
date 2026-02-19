package graphapi

import "github.com/theopenlane/core/pkg/gala"

const (
	// defaultMaxWorkers is the default number of workers when the pool was not created on server startup
	defaultMaxWorkers = 200
)

// withPool returns the existing pool or creates a new one if it does not exist to be used in queries
func (r *queryResolver) withPool() *gala.Pool {
	if r.pool != nil {
		return r.pool
	}

	r.pool = gala.NewPool(
		gala.WithWorkers(defaultMaxWorkers),
		gala.WithPoolName("graphapi-worker-pool"),
	)

	return r.pool
}

// withPool returns the existing pool or creates a new one if it does not exist to be used in mutations
// note that transactions can not be used when using a pool, so this is only used for non-transactional mutations
func (r *mutationResolver) withPool() *gala.Pool {
	if r.pool != nil {
		return r.pool
	}

	r.pool = gala.NewPool(
		gala.WithWorkers(defaultMaxWorkers),
		gala.WithPoolName("graphapi-worker-pool"),
	)

	return r.pool
}
