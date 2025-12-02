package soiree

import "sync"

var (
	registry sync.Map
)

// register adds an event pool to the global registry
func register(pool *EventPool) {
	if pool != nil {
		registry.Store(pool, struct{}{})
	}
}

// deregister removes an event pool from the global registry
func deregister(pool *EventPool) {
	registry.Delete(pool)
}

// ShutdownAll gracefully closes all registered event pools
func ShutdownAll() error {
	var err error

	registry.Range(func(key, _ any) bool {
		pool := key.(*EventPool)

		if closeErr := pool.Close(); closeErr != nil && err == nil {
			err = closeErr
		}

		registry.Delete(key)

		return true
	})

	return err
}
