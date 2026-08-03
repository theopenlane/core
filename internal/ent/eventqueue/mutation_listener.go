package eventqueue

import (
	"context"

	"github.com/samber/do/v2"
	"github.com/samber/lo"

	generated "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// Invocation bundles the standard dependencies handed to a mutation listener:
// the restored event context seeded with the ent client, the resolved client and
// entity identifier, and the dispatch-time injector for additional dependencies
type Invocation struct {
	// Context is the restored event context seeded with the ent client
	Context context.Context
	// Client is the ent client resolved from the gala injector
	Client *generated.Client
	// EntityID is the mutated entity identifier
	EntityID string
	// Envelope is the envelope being processed
	Envelope gala.Envelope
	// Injector provides typed dependency lookup via samber/do
	Injector do.Injector
}

// MutationListener declares one mutation listener bound to a schema mutation topic.
// Registration derives the topic from Concern + Schema and the handler runs behind
// the standard preamble: client resolution, entity-id resolution, and field gating.
// Listeners that operate purely on payload or properties without a resolvable
// entity id should register a raw gala listener instead
type MutationListener struct {
	// Concern selects the mutation topic namespace; empty means MutationConcernDirect
	Concern MutationConcern
	// Schema is the ent schema type name whose mutations the listener observes
	Schema string
	// Name is the stable listener name
	Name string
	// Operations optionally scopes listener interest to specific mutation operations
	Operations []string
	// Fields optionally gates handling on at least one of these fields having changed
	Fields []string
	// Handle is invoked with the standard invocation bundle and the mutation payload
	Handle func(Invocation, MutationGalaPayload) error
}

// RegisterMutationListeners registers mutation listeners with the standard preamble applied
func RegisterMutationListeners(g *gala.Gala, listeners ...MutationListener) ([]gala.ListenerID, error) {
	ids := make([]gala.ListenerID, 0, len(listeners))

	for _, listener := range listeners {
		concern := listener.Concern
		if concern == "" {
			concern = MutationConcernDirect
		}

		registered, err := gala.Register(g, gala.Definition[MutationGalaPayload]{
			Topic:      MutationTopic(concern, listener.Schema),
			Name:       listener.Name,
			Operations: listener.Operations,
			Handle:     mutationHandler(listener),
		})
		if err != nil {
			return nil, err
		}

		ids = append(ids, registered...)
	}

	return ids, nil
}

// mutationHandler wraps a mutation listener handler with the standard preamble
func mutationHandler(listener MutationListener) gala.Handler[MutationGalaPayload] {
	return func(ctx gala.HandlerContext, payload MutationGalaPayload) error {
		if len(listener.Fields) > 0 && !lo.SomeBy(listener.Fields, func(field string) bool {
			return MutationFieldChanged(payload, field)
		}) {
			return nil
		}

		ctx, client, ok := ClientFromHandler(ctx)
		if !ok {
			return nil
		}

		entityID, ok := MutationEntityID(payload, ctx.Envelope.Headers.Properties)
		if !ok {
			logx.FromContext(ctx.Context).Debug().Str("listener", listener.Name).Str("topic", string(ctx.Envelope.Topic)).Msg("mutation listener skipped: entity id unresolved")

			return nil
		}

		return listener.Handle(Invocation{
			Context:  ctx.Context,
			Client:   client,
			EntityID: entityID,
			Envelope: ctx.Envelope,
			Injector: ctx.Injector,
		}, payload)
	}
}
