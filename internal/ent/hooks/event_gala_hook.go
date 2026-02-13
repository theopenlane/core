package hooks

import (
	"context"
	"reflect"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/events"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/workflowgenerated"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/entx"
)

// EmitGalaEventHook returns a hook that emits Gala mutation envelopes after mutations.
func EmitGalaEventHook(galaEmitter *GalaEmitter) ent.Hook {
	if galaEmitter == nil {
		return func(next ent.Mutator) ent.Mutator { return next }
	}

	return hook.If(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, mutation ent.Mutation) (ent.Value, error) {
			// Soft delete emits are handled by the mirrored update mutation path.
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
				topicName := mutation.Type()
				if !galaEmitter.shouldDispatch(topicName) {
					return
				}

				runtime := galaEmitter.runtime()
				if runtime == nil {
					if galaEmitter.failOnEnqueueError {
						logx.FromContext(ctx).Error().Str("topic", topicName).Msg("gala mutation dispatch unavailable")
					}

					return
				}

				metadata := eventqueue.NewMutationGalaMetadata(eventID.ID, payload)
				if galaErr := enqueueGalaMutation(ctx, runtime, topicName, payload, metadata); galaErr != nil && galaEmitter.failOnEnqueueError {
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
		galaEmitter.emitGalaEventOn(),
	)
}

// emitGalaEventOn reports whether the mutation topic is configured for Gala dispatch.
func (g *GalaEmitter) emitGalaEventOn() func(context.Context, entgen.Mutation) bool {
	return func(_ context.Context, m entgen.Mutation) bool { //nolint:revive
		if m == nil {
			return false
		}

		entity := m.Type()
		if entity == "" {
			return false
		}

		return g.shouldDispatch(entity)
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
