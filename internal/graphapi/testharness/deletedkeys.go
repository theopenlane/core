//go:build test

package testharness

import "sync"

// DeletedKeys is a concurrency safe set of the storage keys removed from object storage, the gala
// workers deleting them run on their own goroutines while the test asserts on the main one
type DeletedKeys struct {
	mu   sync.Mutex
	keys map[string]struct{}
}

// NewDeletedKeys returns an empty set
func NewDeletedKeys() *DeletedKeys {
	return &DeletedKeys{keys: map[string]struct{}{}}
}

// Add records a storage key that was deleted
func (d *DeletedKeys) Add(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.keys[key] = struct{}{}
}

// Has reports whether the given storage key was deleted
func (d *DeletedKeys) Has(key string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, ok := d.keys[key]

	return ok
}
