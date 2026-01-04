package soiree

import "sync"

var (
	registry sync.Map
)

// register adds an event bus to the global registry
func register(bus *EventBus) {
	if bus != nil {
		registry.Store(bus, struct{}{})
	}
}

// deregister removes an event bus from the global registry
func deregister(bus *EventBus) {
	registry.Delete(bus)
}

// ShutdownAll gracefully closes all registered event buses
func ShutdownAll() error {
	var err error

	registry.Range(func(key, _ any) bool {
		bus := key.(*EventBus)

		if closeErr := bus.Close(); closeErr != nil && err == nil {
			err = closeErr
		}

		registry.Delete(key)

		return true
	})

	return err
}
