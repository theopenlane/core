package graphapi

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/pkg/gala"
)

var listenerTestQueueSeq atomic.Uint64

// ListenerTestSetup contains the per-test gala runtime for listener integration tests
type ListenerTestSetup struct {
	Runtime *gala.Gala
}

// Teardown stops workers and releases connections
func (s *ListenerTestSetup) Teardown() {
	if s == nil || s.Runtime == nil {
		return
	}

	_ = s.Runtime.StopWorkers(context.Background())
	_ = s.Runtime.Close()
}

// SetupListenerRuntime creates a durable gala runtime with the given listener registrations
// wired to the shared ent client, so a real mutation drives the emit hook, gate, caller
// resolution, and handler exactly as production dispatch does. The runtime and client are
// provided to the injector along with any listener-specific deps; call Teardown on the
// returned setup when the test completes
func SetupListenerRuntime(ctx context.Context, client *generated.Client, connectionURI string, registrations []gala.Registration, deps ...gala.AttachOption) (*ListenerTestSetup, error) {
	if client == nil {
		return nil, ErrClientRequired
	}

	if connectionURI == "" {
		return nil, ErrConnectionURIRequired
	}

	queueName := fmt.Sprintf("listener_test_%d", listenerTestQueueSeq.Add(1))

	runtime, err := gala.NewGala(ctx, gala.Config{
		DispatchMode:      gala.DispatchModeDurable,
		ConnectionURI:     connectionURI,
		QueueName:         queueName,
		WorkerCount:       defaultWorkerCount,
		RunMigrations:     true,
		FetchCooldown:     time.Millisecond,
		FetchPollInterval: defaultFetchPoll,
	})
	if err != nil {
		return nil, err
	}

	client.Use(hooks.EmitGalaEventHook(runtime))

	if _, err := gala.Register(runtime, registrations...); err != nil {
		_ = runtime.Close()

		return nil, err
	}

	opts := append([]gala.AttachOption{
		gala.WithValue(runtime),
		gala.WithValue(client),
		gala.WithRestoredValue("ent_client", generated.NewContext),
	}, deps...)

	if err := runtime.Attach(opts...); err != nil {
		_ = runtime.Close()

		return nil, err
	}

	if err := runtime.StartWorkers(ctx); err != nil {
		_ = runtime.Close()

		return nil, err
	}

	return &ListenerTestSetup{
		Runtime: runtime,
	}, nil
}
