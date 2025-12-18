package graphapi

import "github.com/theopenlane/core/pkg/events/soiree"

const (
	// defaultMaxWorkers is the default number of workers in the pond pool when the pool was not created on server startup
	defaultMaxWorkers = 10
)

// WithPool returns the existing pool or creates a new one if it does not exist to be used in queries
func (r *queryResolver) withPool() *soiree.PondPool {
	if r.pool != nil {
		return r.pool
	}

	r.pool = soiree.NewPondPool(soiree.WithMaxWorkers(defaultMaxWorkers))

	return r.pool
}

// WithPool returns the existing pool or creates a new one if it does not exist to be used in mutations
// note that transactions can not be used when using a pool, so this is only used for non-transactional mutations
func (r *mutationResolver) withPool() *soiree.PondPool {
	if r.pool != nil {
		return r.pool
	}

	r.pool = soiree.NewPondPool(soiree.WithMaxWorkers(defaultMaxWorkers))

	return r.pool
}
