package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"entgo.io/ent"

	"github.com/99designs/gqlgen/graphql"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/internal/ent/events"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/entx"
)

// EmitEventHook returns a hook that emits events after mutations
func EmitEventHook(e *Eventer) ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, mutation ent.Mutation) (ent.Value, error) {
			// if this is a soft delete, skip emitting events, it will be handled by the duplicate update mutation that is triggered
			// otherwise, you'll get double events for soft deletes
			if entx.CheckIsSoftDeleteType(ctx, mutation.Type()) {
				return next.Mutate(ctx, mutation)
			}

			retVal, err := next.Mutate(ctx, mutation)
			if err != nil {
				return nil, err
			}

			// determine the operation type
			op := getOperation(ctx, mutation)

			// Delete operations return an int of the number of rows deleted
			// so we do not want to skip emitting events for those operations
			if op != SoftDeleteOne && reflect.TypeOf(retVal).Kind() == reflect.Int {
				logx.FromContext(ctx).Debug().Interface("value", retVal).Msgf("mutation of type %s returned an int, skipping event emission", op)
				return retVal, err
			}

			emit := func() {
				eventID := &EventID{}
				if op == SoftDeleteOne {
					eventID, err = parseSoftDeleteEventID(ctx, mutation)
					if err != nil {
						logx.FromContext(ctx).Info().Err(err).Msg("failed to parse event ID for soft delete, skipping event emission")

						return
					}
				} else {
					eventID, err = parseEventID(retVal)
					if err != nil {
						logx.FromContext(ctx).Error().Err(err).Msg("failed to parse event ID, skipping event emission")

						return
					}
				}

				if eventID == nil || eventID.ID == "" {
					logx.FromContext(ctx).Error().Err(err).Msg("Event ID is nil or empty, cannot emit event")
					return
				}

				// Create a child logger for concurrency safety
				logger := log.Logger.With().Logger()
				logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
					return c.Str("mutation_id", eventID.ID)
				})

				props := soiree.NewProperties()
				props.Set("ID", eventID.ID)
				addMutationFields(props, mutation)

				payload := &events.MutationPayload{
					Mutation:  mutation,
					Operation: op,
					EntityID:  eventID.ID,
				}

				var emitterClient any
				if e.Emitter != nil {
					emitterClient = e.Emitter.Client()
					if client, ok := emitterClient.(*entgen.Client); ok {
						payload.Client = client
					}
				}

				topic := mutationTopic(mutation.Type())

				event := soiree.NewBaseEvent(topic.Name(), payload)
				event.SetProperties(props)
				event.SetContext(context.WithoutCancel(ctx))
				if payload.Client != nil {
					event.SetClient(payload.Client)
				} else if emitterClient != nil {
					event.SetClient(emitterClient)
				}

				if e.Emitter != nil {
					// fire-and-forget; listeners drain the returned channel
					e.Emitter.Emit(topic.Name(), event)
				}
			}

			if tx := transactionFromContext(ctx); tx != nil {
				tx.OnCommit(func(next entgen.Committer) entgen.Committer {
					return entgen.CommitFunc(func(ctx context.Context, tx *entgen.Tx) error {
						err := next.Commit(ctx, tx)
						if err == nil {
							defer emit()
						}

						return err
					})
				})
			} else {
				defer emit()
			}

			return retVal, err
		})
	},
		e.emitEventOn(),
	)
}

// EventID represents the ID structure used in events
type EventID struct {
	ID string `json:"id,omitempty"`
}

// parseEventID extracts the EventID from the returned value of a mutation
func parseEventID(retVal ent.Value) (*EventID, error) {
	out, err := json.Marshal(retVal)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal return value")
		return nil, fmt.Errorf("failed to fetch organization from subscription: %w", err)
	}

	event := EventID{}
	if err := json.Unmarshal(out, &event); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal return value")
		return nil, err
	}

	return &event, nil
}

// parseSoftDeleteEventID extracts the EventID from a soft delete mutation
func parseSoftDeleteEventID(ctx context.Context, mutation ent.Mutation) (*EventID, error) {
	mut, ok := mutation.(utils.GenericMutation)
	if !ok {
		return nil, ErrUnableToDetermineEventID
	}

	ids := getMutationIDs(ctx, mut)
	if len(ids) == 0 || ids[0] == "" {
		return nil, ErrUnableToDetermineEventID
	}

	if len(ids) > 1 {
		logx.FromContext(ctx).Warn().Strs("mutation_ids", ids).Msg("Soft delete mutation returned multiple IDs")
	}

	return &EventID{ID: ids[0]}, nil
}

// getOperation determines the operation type from the context and mutation
func getOperation(ctx context.Context, mutation ent.Mutation) string {
	// check the graphql operation context for the operation name
	if graphql.HasOperationContext(ctx) {
		opCtx := graphql.GetOperationContext(ctx)
		if opCtx != nil {
			if opCtx.OperationName == "DeleteOrganization" && mutation.Type() == entgen.TypeOrganization {
				return SoftDeleteOne
			}
		}
	}

	return mutation.Op().String()
}

// emitEventOn determines whether to emit events for a given mutation
func (e *Eventer) emitEventOn() func(context.Context, entgen.Mutation) bool {
	return func(ctx context.Context, m entgen.Mutation) bool { //nolint:revive
		if e == nil || m == nil {
			return false
		}

		entity := m.Type()
		if entity == "" {
			return false
		}

		// Prefer the live pool state so dynamically registered listeners are honoured even when they bypass
		// Eventer bookkeeping (e.g. direct EventBus.On calls in tests)
		if e.Emitter != nil && e.Emitter.InterestedIn(entity) {
			return true
		}

		if e.listeners == nil {
			return false
		}

		// Listener registration drives emission: if no subscribers, we avoid creating events altogether
		listeners, ok := e.listeners[entity]

		return ok && len(listeners) > 0
	}
}

const (
	SoftDeleteOne = "SoftDeleteOne"
)

// RegisterListeners registers all listeners on the Eventer with the emitter
func RegisterListeners(e *Eventer) error {
	if e.Emitter == nil {
		log.Error().Msg("Emitter is nil on Eventer, cannot register listeners")

		return ErrFailedToRegisterListener
	}

	total := 0
	for _, entries := range e.listeners {
		total += len(entries)
	}

	if total == 0 {
		return nil
	}

	bindings := make([]soiree.ListenerBinding, 0, total)
	for _, entries := range e.listeners {
		bindings = append(bindings, entries...)
	}

	if _, err := e.Emitter.RegisterListeners(bindings...); err != nil {
		log.Error().Err(err).Msg("failed to register listeners")
		return err
	}

	return nil
}

// addMutationFields adds all fields from the mutation to the event properties
func addMutationFields(props soiree.Properties, mutation ent.Mutation) {
	if props == nil || mutation == nil {
		return
	}

	for _, field := range mutation.Fields() {
		if value, ok := mutation.Field(field); ok {
			props.Set(field, value)
		}
	}
}
