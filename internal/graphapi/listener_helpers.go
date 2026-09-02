package graphapi

import (
	"context"

	"github.com/theopenlane/core/v2/pkg/gala"
)

// ListenerTestSetup tracks temporary listeners on the shared test runtime
type ListenerTestSetup struct {
	Runtime     *gala.Gala
	listenerIDs []gala.ListenerID
}

// Teardown detaches listeners and purges their durable jobs
func (s *ListenerTestSetup) Teardown() {
	if s == nil || s.Runtime == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), gala.DefaultSoftStopTimeout)
	defer cancel()

	if err := s.Runtime.RemoveListeners(ctx, s.listenerIDs...); err != nil {
		panic(err)
	}
}

// SetupListenerRuntime registers temporary listeners on the shared runtime
// Registry changes are visible through the existing Ent hook
func SetupListenerRuntime(runtime *gala.Gala, registrations []gala.Registration) (*ListenerTestSetup, error) {
	if runtime == nil {
		return nil, gala.ErrGalaRequired
	}

	listenerIDs, err := gala.Register(runtime, registrations...)
	if err != nil {
		_ = runtime.RemoveListeners(context.Background(), listenerIDs...)

		return nil, err
	}

	return &ListenerTestSetup{
		Runtime:     runtime,
		listenerIDs: listenerIDs,
	}, nil
}
