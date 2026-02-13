package eventqueue

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/riverqueue/river"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/events"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/events/soiree"
)

var (
	// ErrEmitterUnavailable indicates the worker cannot dispatch because the event bus is unavailable.
	ErrEmitterUnavailable = errors.New("mutation dispatch worker: emitter unavailable")
	// ErrTopicRequired indicates the dispatched job did not contain a valid topic.
	ErrTopicRequired = errors.New("mutation dispatch worker: topic is required")
)

// MutationDispatchWorker dequeues mutation envelopes from River and re-emits them to Soiree.
type MutationDispatchWorker struct {
	river.WorkerDefaults[MutationDispatchArgs]

	busProvider func() *soiree.EventBus
}

// NewMutationDispatchWorker constructs a worker that resolves the active event bus at runtime.
func NewMutationDispatchWorker(busProvider func() *soiree.EventBus) *MutationDispatchWorker {
	return &MutationDispatchWorker{
		busProvider: busProvider,
	}
}

// Work processes one mutation dispatch job and blocks until listener execution completes.
func (w *MutationDispatchWorker) Work(ctx context.Context, job *river.Job[MutationDispatchArgs]) error {
	if w.busProvider == nil {
		return ErrEmitterUnavailable
	}

	bus := w.busProvider()
	if bus == nil {
		return ErrEmitterUnavailable
	}

	args := job.Args
	if strings.TrimSpace(args.Topic) == "" {
		return ErrTopicRequired
	}

	payload := &events.MutationPayload{
		MutationType:    args.MutationType,
		Operation:       args.Operation,
		EntityID:        args.EntityID,
		ChangedFields:   append([]string(nil), args.ChangedFields...),
		ClearedFields:   append([]string(nil), args.ClearedFields...),
		ChangedEdges:    append([]string(nil), args.ChangedEdges...),
		AddedIDs:        events.CloneStringSliceMap(args.AddedIDs),
		RemovedIDs:      events.CloneStringSliceMap(args.RemovedIDs),
		ProposedChanges: events.CloneAnyMap(args.ProposedChanges),
	}

	if client, ok := bus.Client().(*generated.Client); ok {
		payload.Client = client
	}

	event := soiree.NewBaseEvent(args.Topic, payload)

	props := soiree.NewProperties()
	for key, value := range args.Properties {
		if strings.TrimSpace(key) == "" {
			continue
		}

		props.Set(key, value)
	}

	if args.EntityID != "" {
		props.Set("ID", args.EntityID)
	}

	if args.EventID != "" {
		props.Set(soiree.PropertyEventID, args.EventID)
	}

	if args.MutationType != "" {
		props.Set("mutation_type", args.MutationType)
	}

	event.SetProperties(props)

	emitCtx := context.WithoutCancel(ctx)
	if authCtx := args.Auth; authCtx != nil {
		if au := authCtx.ToAuthenticatedUser(); au != nil {
			emitCtx = auth.WithAuthenticatedUser(emitCtx, au)
		}
	}

	event.SetContext(emitCtx)

	if payload.Client != nil {
		event.SetClient(payload.Client)
	} else if client := bus.Client(); client != nil {
		event.SetClient(client)
	}

	errCh := bus.Emit(args.Topic, event)
	var joined error

	for emitErr := range errCh {
		if emitErr != nil {
			joined = errors.Join(joined, emitErr)
		}
	}

	if joined != nil {
		return fmt.Errorf("mutation dispatch worker: emit failed: %w", joined)
	}

	return nil
}
