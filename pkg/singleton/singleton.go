package singleton

import "sync/atomic"

// Value holds a process-wide default instance of T
type Value[T any] struct {
	ptr atomic.Pointer[T]
}

// Set registers the process-wide default instance
func (v *Value[T]) Set(instance *T) {
	v.ptr.Store(instance)
}

// Get returns the process-wide default instance, or nil when none is registered
func (v *Value[T]) Get() *T {
	return v.ptr.Load()
}
