package soiree

import (
	"sort"
	"sync"
)

// Topic represents an event channel to which listeners can subscribe
type Topic struct {
	// Name signifies the topic's unique identifier
	Name string
	// Mutex for concurrent access to the topic
	mu sync.RWMutex
	// Map of listeners indexed by their ID
	listeners map[string]*listenerItem
	// Sorted list of listener IDs for priority-based iteration
	sortedListenerIDs []string
}

// NewTopic creates a new Topic
func NewTopic() *Topic {
	return &Topic{
		listeners: make(map[string]*listenerItem),
	}
}

// addSortedListenerID inserts a listener ID into the sorted juicy slice at the correct position
func (t *Topic) addSortedListenerID(id string, priority Priority) {
	index := sort.Search(len(t.sortedListenerIDs), func(i int) bool {
		return t.listeners[t.sortedListenerIDs[i]].priority <= priority
	})
	// Extend the slice by one element
	t.sortedListenerIDs = append(t.sortedListenerIDs, "")
	// Move the higher priority elements up one slot to make room for the new ID
	copy(t.sortedListenerIDs[index+1:], t.sortedListenerIDs[index:])
	// Set the new ID in the correct slot
	t.sortedListenerIDs[index] = id
}

// removeSortedListenerID removes a listener ID from the sorted slice
func (t *Topic) removeSortedListenerID(id string) {
	for i, listenerID := range t.sortedListenerIDs {
		if listenerID == id {
			t.sortedListenerIDs = append(t.sortedListenerIDs[:i], t.sortedListenerIDs[i+1:]...)
			break
		}
	}
}

// AddListener adds a new listener to the topic with a specified priority and returns an identifier for the listener
func (t *Topic) AddListener(id string, listener Listener, opts ...ListenerOption) {
	t.mu.Lock()
	defer t.mu.Unlock()

	item := &listenerItem{
		listener: listener,
		priority: Normal, // Default priority if none is specified
	}

	for _, opt := range opts {
		opt(item)
	}

	t.listeners[id] = item
	t.addSortedListenerID(id, item.priority)
}

// RemoveListener removes a listener from the topic using its identifier
func (t *Topic) RemoveListener(id string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, ok := t.listeners[id]; !ok {
		return ErrListenerNotFound
	}

	delete(t.listeners, id)
	t.removeSortedListenerID(id)

	return nil
}

// Trigger calls all listeners of the topic with the event
func (t *Topic) Trigger(event Event) []error {
	t.mu.RLock()

	defer t.mu.RUnlock()

	var errs []error

	for _, id := range t.sortedListenerIDs {
		item, ok := t.listeners[id]
		if !ok {
			continue // Listener was removed; skip it
		}

		ctx := newEventContext(event)

		if err := item.call(ctx); err != nil {
			errs = append(errs, err)
		}

		if ctx.Event().IsAborted() {
			break // Stop notifying listeners if the event is aborted
		}
	}

	return errs
}

// HasListeners reports whether the topic currently has listeners registered.
func (t *Topic) HasListeners() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return len(t.listeners) > 0
}
