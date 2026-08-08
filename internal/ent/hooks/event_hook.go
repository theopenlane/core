package hooks

import (
	"context"
	"fmt"
	"reflect"

	"entgo.io/ent"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/entityops"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/entx"
)

// EmitGalaEventHook returns a hook that emits Gala mutation envelopes after mutations.
// Runtimes are deduplicated once at installation; a mutation fans out to every concern
// topic each runtime has an interested listener for
func EmitGalaEventHook(runtimes ...*gala.Gala) ent.Hook {
	galaRuntimes := lo.Uniq(lo.Compact(runtimes))

	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, mutation ent.Mutation) (ent.Value, error) {
			if entx.CheckIsSoftDeleteType(ctx, mutation.Type()) {
				return next.Mutate(ctx, mutation)
			}

			ctx = gala.WithSkipEventEmission(ctx)

			// classify before mutating: the soft-delete mixin rewrites the mutation op to
			// an update in place, so the original delete op is only observable here
			op := mutationOperation(ctx, mutation)

			oldValues := snapshotOldValues(ctx, mutation, galaRuntimes)

			retVal, err := next.Mutate(ctx, mutation)
			if err != nil {
				return nil, err
			}

			if gala.ShouldSkipEventEmission(ctx) {
				return retVal, err
			}

			if op != gala.SoftDeleteOne && retVal != nil && reflect.TypeOf(retVal).Kind() == reflect.Int {
				return retVal, err
			}

			topicName := mutation.Type()
			if topicName == "" {
				return retVal, err
			}

			emit := func() {
				if !anyRuntimeInterested(galaRuntimes, topicName, op) {
					return
				}

				entityID, idErr := mutationEventEntityID(ctx, mutation, op, retVal)
				if idErr != nil || entityID == "" {
					logx.FromContext(ctx).Error().Err(idErr).Str("mutation_type", topicName).Msg("failed to resolve mutation event id, skipping gala emission")

					return
				}

				payload := entityops.MutationPayload{
					MutationType: topicName,
					Operation:    op,
					EntityID:     entityID,
					ChangeSet:    entityops.ChangeSetFromMutation(mutation),
				}
				payload.OldValues = oldValues

				headers := entityops.MutationHeaders(payload)

				// detach cancellation for best-effort dispatch after commit
				dispatchCtx := context.WithoutCancel(ctx)

				for _, runtime := range galaRuntimes {
					for _, topic := range mutationConcernTopics(topicName) {
						if !runtime.InterestedIn(topic, op) {
							continue
						}

						if _, galaErr := runtime.Emit(dispatchCtx, topic, payload,
							gala.WithHeaders(headers),
							gala.WithEventID(gala.EventID(entityID))); galaErr != nil {
							logx.FromContext(ctx).Error().Err(fmt.Errorf("%w: emit: %w", ErrGalaMutationEnqueueFailed, galaErr)).Str("topic", string(topic)).Msg("gala mutation dispatch failed")
						}
					}
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
	}
}

// mutationConcernTopics returns the concern topic names a schema mutation fans out to
func mutationConcernTopics(schemaType string) [3]gala.TopicName {
	return [3]gala.TopicName{
		gala.MutationTopicName(gala.MutationConcernDirect, schemaType),
		gala.MutationTopicName(gala.MutationConcernWorkflow, schemaType),
		gala.MutationTopicName(gala.MutationConcernNotification, schemaType),
	}
}

// anyRuntimeInterested reports whether any runtime has an interested listener on any of
// the schema's concern topics for the operation
func anyRuntimeInterested(runtimes []*gala.Gala, schemaType, operation string) bool {
	for _, runtime := range runtimes {
		for _, topic := range mutationConcernTopics(schemaType) {
			if runtime.InterestedIn(topic, operation) {
				return true
			}
		}
	}

	return false
}

// snapshotOldValues captures pre-update field values before the mutation applies, while the
// database still holds the prior row. Capture is limited to single-row updates whose topics
// have at least one interested gala listener, so unobserved mutations cost nothing; ent
// caches the loaded old row on the mutation, so hooks that already called OldField share it
func snapshotOldValues(ctx context.Context, mutation ent.Mutation, runtimes []*gala.Gala) map[string]any {
	source, ok := mutation.(entityops.OldValueSource)
	if !ok || !mutation.Op().Is(ent.OpUpdateOne) {
		return nil
	}

	if !anyRuntimeInterested(runtimes, mutation.Type(), mutation.Op().String()) {
		return nil
	}

	changed := entityops.ChangeSetFromMutation(mutation).ChangedFields

	return entityops.BuildOldValues(ctx, source, changed)
}
