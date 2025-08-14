package reconciler

import (
	"context"

	"github.com/theopenlane/core/pkg/reconciler"
)

// Wrapper is a concrete type that wraps the Reconciler interface
// This is needed for ent dependency injection which requires concrete types
type Wrapper struct {
	impl reconciler.Reconciler
}

// NewWrapper creates a new wrapper around a reconciler implementation
func NewWrapper(r reconciler.Reconciler) *Wrapper {
	return &Wrapper{impl: r}
}

// Reconcile implements the reconciler.Reconciler interface
func (w *Wrapper) Reconcile(ctx context.Context) error {
	if w.impl == nil {
		return nil // no-op if not configured
	}
	return w.impl.Reconcile(ctx)
}

// GetImplementation returns the underlying reconciler implementation
func (w *Wrapper) GetImplementation() reconciler.Reconciler {
	return w.impl
}
