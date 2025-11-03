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
	"github.com/samber/lo"

	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/entx"
)

type EventID struct {
	ID string `json:"id,omitempty"`
}

// parseEventID parses the event ID from the return value of an ent mutation
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

// parseSoftDeleteEventID parses the event ID from a soft delete organization mutation by casting the mutation to an organization mutation
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

// EmitEventHook emits an event to the event pool when a mutation is performed
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
				logx.FromContext(ctx).Debug().Interface("value", retVal).Msgf("mutation of type %s returned an int, skipping event emission", op)
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

				topic := soiree.MutationTopic(mutation.Type(), op)
				event := soiree.NewBaseEvent(topic.Name(), mutation)
				properties := soiree.NewProperties()
				properties.Set("ID", eventID.ID)

				if e != nil {
					e.applyPropertyExtractors(ctx, topic.Name(), mutation, properties)
				}

				event.SetProperties(properties)

				zerolog.Ctx(ctx).UpdateContext(func(c zerolog.Context) zerolog.Context {
					return c.Str("mutation_id", eventID.ID)
				})

				event.SetContext(context.WithoutCancel(ctx))
				event.SetClient(e.Emitter.GetClient())
				soiree.EmitTopic(e.Emitter, topic, soiree.Event(event))
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
		emitEventOn(),
	)
}

// getOperation gets the operation from the context or mutation
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

// emitEventOn is a function that returns a function that checks if an event should be emitted
// based on the mutation type and operation and fields that were updated
func emitEventOn() func(context.Context, entgen.Mutation) bool {
	return func(ctx context.Context, m entgen.Mutation) bool {
		op := getOperation(ctx, m)

		rules, ok := mutationRules[m.Type()]
		if !ok {
			return false
		}

		for _, rule := range rules {
			if len(rule.Operations) > 0 && !lo.Contains(rule.Operations, op) {
				continue
			}

			if len(rule.Fields) > 0 {
				fieldMatch := lo.ContainsBy(rule.Fields, func(field string) bool {
					_, ok := m.Field(field)
					return ok
				})

				if !fieldMatch {
					continue
				}
			}

			if rule.Condition != nil && !rule.Condition(ctx, m) {
				continue
			}

			return true
		}

		return false
	}
}

const (
	SoftDeleteOne = "SoftDeleteOne"
)

type mutationRule struct {
	Operations []string
	Fields     []string
	Condition  func(context.Context, entgen.Mutation) bool
}

var mutationRules = map[string][]mutationRule{
	entgen.TypeOrgSubscription: {
		{Operations: []string{ent.OpCreate.String()}},
	},
	entgen.TypeOrganizationSetting: {
		{Operations: []string{ent.OpUpdateOne.String(), ent.OpUpdate.String()}, Fields: []string{"billing_email", "billing_phone", "billing_address"}},
	},
	entgen.TypeOrganization: {
		{Operations: []string{ent.OpDelete.String(), ent.OpDeleteOne.String(), ent.OpCreate.String(), SoftDeleteOne}},
	},
	entgen.TypeSubscriber: {
		{Operations: []string{ent.OpCreate.String()}},
	},
	entgen.TypeUser: {
		{Operations: []string{ent.OpCreate.String()}},
	},
}

var organizationDeleteOps = []string{ent.OpDelete.String(), ent.OpDeleteOne.String(), SoftDeleteOne}
var organizationSettingOps = []string{ent.OpUpdateOne.String(), ent.OpUpdate.String()}

var mutationListenerBindings = append(
	append(
		[]soiree.ListenerBinding{
			soiree.BindContextListener(soiree.MutationTopic(entgen.TypeOrganization, ent.OpCreate.String()), handleOrganizationCreated),
			soiree.BindContextListener(soiree.MutationTopic(entgen.TypeSubscriber, ent.OpCreate.String()), handleSubscriberCreate),
			soiree.BindContextListener(soiree.MutationTopic(entgen.TypeUser, ent.OpCreate.String()), handleUserCreate),
		},
		lo.Map(organizationSettingOps, func(op string, _ int) soiree.ListenerBinding {
			return soiree.BindContextListener(soiree.MutationTopic(entgen.TypeOrganizationSetting, op), handleOrganizationSettingsUpdateOne)
		})...,
	),
	lo.Map(organizationDeleteOps, func(op string, _ int) soiree.ListenerBinding {
		return soiree.BindContextListener(soiree.MutationTopic(entgen.TypeOrganization, op), handleOrganizationDelete)
	})...,
)

// RegisterListeners is currently used to globally register what listeners get applied on the entdb client
func RegisterListeners(e *Eventer) error {
	if e.Emitter == nil {
		log.Error().Msg("Emitter is nil on Eventer, cannot register listeners")

		return ErrFailedToRegisterListener
	}

	allBindings := make([]soiree.ListenerBinding, 0, len(mutationListenerBindings)+len(e.listenerBindings))
	allBindings = append(allBindings, mutationListenerBindings...)
	allBindings = append(allBindings, e.listenerBindings...)

	if len(allBindings) == 0 {
		return nil
	}

	if _, err := e.Emitter.RegisterListeners(allBindings...); err != nil {
		log.Error().Err(err).Msg("failed to register listeners")
		return err
	}

	return nil
}
