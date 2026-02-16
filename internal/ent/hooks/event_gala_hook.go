package hooks

import (
	"context"
	"reflect"

	"entgo.io/ent"

	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/events"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/workflowgenerated"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/entx"
)

// EmitGalaEventHook returns a hook that emits Gala mutation envelopes after mutations.
func EmitGalaEventHook(galaProvider func() *gala.Gala, failOnEnqueueError bool) ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, mutation ent.Mutation) (ent.Value, error) {
			if entx.CheckIsSoftDeleteType(ctx, mutation.Type()) {
				return next.Mutate(ctx, mutation)
			}

			ctx = workflows.WithSkipEventEmission(ctx)

			retVal, err := next.Mutate(ctx, mutation)
			if err != nil {
				return nil, err
			}

			if workflows.ShouldSkipEventEmission(ctx) {
				return retVal, err
			}

			op := getOperation(ctx, mutation)

			if op != SoftDeleteOne && reflect.TypeOf(retVal).Kind() == reflect.Int {
				return retVal, err
			}

			topicName := mutation.Type()
			if topicName == "" {
				return retVal, err
			}

			emit := func() {
				eventID := &EventID{}
				if op == SoftDeleteOne {
					eventID, err = parseSoftDeleteEventID(ctx, mutation)
					if err != nil {
						logx.FromContext(ctx).Info().Err(err).Msg("failed to parse event ID for soft delete, skipping gala emission")

						return
					}
				} else {
					eventID, err = parseEventID(retVal)
					if err != nil {
						logx.FromContext(ctx).Error().Err(err).Msg("failed to parse event ID, skipping gala emission")

						return
					}
				}

				if eventID == nil || eventID.ID == "" {
					logx.FromContext(ctx).Error().Msg("event ID is nil or empty, skipping gala emission")

					return
				}

				payload := newMutationPayloadForDispatch(mutation, op, eventID.ID)
				galaRuntime := galaProvider()
				metadata := eventqueue.NewMutationGalaMetadata(eventID.ID, payload)

				if galaErr := enqueueGalaMutation(ctx, galaRuntime, topicName, payload, metadata); galaErr != nil && failOnEnqueueError {
					logx.FromContext(ctx).Error().Err(galaErr).Str("topic", topicName).Msg("gala mutation dispatch failed")
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
		emitGalaEventOn(galaProvider),
	)
}

// emitGalaEventOn reports whether the mutation topic is eligible for Gala dispatch
func emitGalaEventOn(galaProvider func() *gala.Gala) func(context.Context, entgen.Mutation) bool {
	return func(ctx context.Context, m entgen.Mutation) bool {
		if m == nil {
			return false
		}

		entity := m.Type()
		if entity == "" || galaProvider == nil {
			return false
		}

		g := galaProvider()
		if g == nil {
			return false
		}

		operation := getOperation(ctx, m)

		return g.Registry().InterestedIn(gala.TopicName(entity), operation)
	}
}

// newMutationPayloadForDispatch builds shared mutation payload metadata for asynchronous dispatch hooks.
func newMutationPayloadForDispatch(mutation ent.Mutation, operation, entityID string) *events.MutationPayload {
	changedFields, clearedFields := mutationChangedAndClearedFields(mutation)
	changedEdges, addedIDs, removedIDs := workflowgenerated.ExtractChangedEdges(mutation)
	proposedChanges := mutationProposedChanges(mutation, changedFields, clearedFields)

	return &events.MutationPayload{
		Mutation:        mutation,
		MutationType:    mutation.Type(),
		Operation:       operation,
		EntityID:        entityID,
		ChangedFields:   changedFields,
		ClearedFields:   clearedFields,
		ChangedEdges:    changedEdges,
		AddedIDs:        addedIDs,
		RemovedIDs:      removedIDs,
		ProposedChanges: proposedChanges,
	}
}

// mutationChangedAndClearedFields derives updated/cleared field names from an ent mutation.
func mutationChangedAndClearedFields(mutation ent.Mutation) ([]string, []string) {
	if mutation == nil {
		return nil, nil
	}

	clearedFields := uniqueStrings(mutation.ClearedFields())
	changedFields := append(append([]string(nil), mutation.Fields()...), clearedFields...)

	return uniqueStrings(changedFields), clearedFields
}

// mutationProposedChanges materializes field values (including explicit clears as nil).
func mutationProposedChanges(mutation ent.Mutation, changedFields, clearedFields []string) map[string]any {
	if mutation == nil || len(changedFields) == 0 {
		return nil
	}

	clearedSet := make(map[string]struct{}, len(clearedFields))
	lo.ForEach(clearedFields, func(field string, _ int) {
		if field == "" {
			return
		}

		clearedSet[field] = struct{}{}
	})

	proposed := make(map[string]any, len(changedFields))
	lo.ForEach(changedFields, func(field string, _ int) {
		if field == "" {
			return
		}

		if val, ok := mutation.Field(field); ok {
			proposed[field] = val
			return
		}

		if _, ok := clearedSet[field]; ok {
			proposed[field] = nil
		}
	})

	if len(proposed) == 0 {
		return nil
	}

	return proposed
}

// uniqueStrings returns distinct non-empty values while preserving first-seen order.
func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	out := lo.Uniq(lo.Filter(values, func(value string, _ int) bool { return value != "" }))
	if len(out) == 0 {
		return nil
	}

	return out
}
