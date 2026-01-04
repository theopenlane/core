package soiree

import (
	"sync"
)

// topic represents an event channel to which listeners can subscribe
type topic struct {
	// name signifies the topic's unique identifier
	name string
	// mu provides concurrent access to the topic
	mu sync.RWMutex
	// listeners indexed by their ID
	listeners map[string]*listenerItem
	// listenerIDs maintains registration order
	listenerIDs []string
}

// newTopic creates a new topic
func newTopic() *topic {
	return &topic{
		listeners: make(map[string]*listenerItem),
	}
}

// addListener adds a new listener to the topic
func (t *topic) addListener(id string, listener Listener) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.listeners[id] = &listenerItem{listener: listener}
	t.listenerIDs = append(t.listenerIDs, id)
}

// removeListener removes a listener from the topic using its identifier
func (t *topic) removeListener(id string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, ok := t.listeners[id]; !ok {
		return ErrListenerNotFound
	}

	delete(t.listeners, id)

	for i, lid := range t.listenerIDs {
		if lid == id {
			t.listenerIDs = append(t.listenerIDs[:i], t.listenerIDs[i+1:]...)
			break
		}
	}

	return nil
}

// hasListeners reports whether the topic currently has listeners registered
func (t *topic) hasListeners() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return len(t.listeners) > 0
}
