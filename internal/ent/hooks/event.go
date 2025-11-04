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

	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/entx"
)

type EventID struct {
	ID string `json:"id,omitempty"`
}

func parseEventID(retVal ent.Value) (*EventID, error) {
	out, err := json.Marshal(retVal)
	if err != nil {
		log.Err(err).Msg("Failed to marshal return value")
		return nil, fmt.Errorf("failed to fetch organization from subscription: %w", err)
	}

	event := EventID{}
	if err := json.Unmarshal(out, &event); err != nil {
		log.Err(err).Msg("Failed to unmarshal return value")
		return nil, err
	}

	return &event, nil
}

func parseSoftDeleteEventID(mutation ent.Mutation) (*EventID, error) {
	m, ok := mutation.(*entgen.OrganizationMutation)
	if !ok {
		return nil, ErrUnableToDetermineEventID
	}

	id, ok := m.ID()
	if !ok {
		return nil, ErrUnableToDetermineEventID
	}

	return &EventID{ID: id}, nil
}

func EmitEventHook(e *Eventer) ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, mutation ent.Mutation) (ent.Value, error) {
			retVal, err := next.Mutate(ctx, mutation)
			if err != nil {
				return nil, err
			}

			// determine the operation type
			op := getOperation(ctx, mutation)

			// Delete operations return an int of the number of rows deleted
			// so we do not want to skip emitting events for those operations
			if op != SoftDeleteOne && reflect.TypeOf(retVal).Kind() == reflect.Int {
				zerolog.Ctx(ctx).Debug().Interface("value", retVal).Msgf("mutation of type %s returned an int, skipping event emission", op)
				// TODO: determine if we need to emit events for mutations that return an int
				return retVal, err
			}

			emit := func() {
				eventID := &EventID{}
				if op == SoftDeleteOne {
					eventID, err = parseSoftDeleteEventID(mutation)
					if err != nil {
						log.Err(err).Msg("Failed to parse soft delete event ID")

						return
					}
				} else {
					eventID, err = parseEventID(retVal)
					if err != nil {
						log.Err(err).Msg("Failed to parse event ID")
						return
					}
				}

				if eventID == nil || eventID.ID == "" {
					log.Err(ErrUnableToDetermineEventID).Msg("Event ID is nil or empty, cannot emit event")
					return
				}

				zerolog.Ctx(ctx).UpdateContext(func(c zerolog.Context) zerolog.Context {
					return c.Str("mutation_id", eventID.ID)
				})

				props := soiree.NewProperties()
				props.Set("ID", eventID.ID)
				addMutationFields(props, mutation)

				payload := &MutationPayload{
					Mutation:  mutation,
					Operation: op,
					EntityID:  eventID.ID,
				}

				var emitterClient any
				if e.Emitter != nil {
					emitterClient = e.Emitter.GetClient()
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

func getOperation(ctx context.Context, mutation ent.Mutation) string {
	// determine if this is a soft delete operation
	// this isn't in the context when we reach here, but incase it is in the future, we check
	if entx.CheckIsSoftDelete(ctx) {
		return SoftDeleteOne
	}

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

func (e *Eventer) emitEventOn() func(context.Context, entgen.Mutation) bool {
	return func(ctx context.Context, m entgen.Mutation) bool {
		if e == nil {
			return false
		}

		if e.entities == nil {
			return false
		}

		_, ok := e.entities[m.Type()]
		return ok
	}
}

const (
	SoftDeleteOne = "SoftDeleteOne"
)

func RegisterListeners(e *Eventer) error {
	if e.Emitter == nil {
		log.Error().Msg("Emitter is nil on Eventer, cannot register listeners")

		return ErrFailedToRegisterListener
	}

	bindings := make([]soiree.ListenerBinding, 0, len(e.listeners))
	for _, listener := range e.listeners {
		bindings = append(bindings, listener.binding)
	}

	if len(bindings) == 0 {
		return nil
	}

	if _, err := e.Emitter.RegisterListeners(bindings...); err != nil {
		log.Error().Err(err).Msg("failed to register listeners")
		return err
	}

	return nil
}

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
