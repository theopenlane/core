package hooks

import (
	"context"
	"reflect"

	"entgo.io/ent"
	"github.com/samber/lo"
	"github.com/theopenlane/entx"

	"github.com/theopenlane/core/internal/ent/entityops"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// EmitGalaEventHook returns a hook that emits Gala mutation envelopes after mutations.
// Runtimes are deduplicated once at installation; a mutation fans out to every concern
// topic each runtime has an interested listener for, one envelope per mutated row
func EmitGalaEventHook(runtimes ...*gala.Gala) ent.Hook {
	galaRuntimes := lo.Uniq(lo.Compact(runtimes))

	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, mutation ent.Mutation) (ent.Value, error) {
			// the rewritten update runs inside the delete pass, which owns emission
			if entx.CheckIsSoftDeleteType(ctx, mutation.Type()) {
				return next.Mutate(ctx, mutation)
			}

			op := mutation.Op().String()
			if softDeleteClassified(ctx, mutation) {
				op = entityops.OpSoftDelete
			}

			ctx = entityops.WithEmissionVeto(ctx)

			changeSet := entityops.ChangeSetFromMutation(mutation)
			ids, oldValues, snapshotErr := snapshotMutation(ctx, mutation, galaRuntimes, op, changeSet)
			if snapshotErr != nil {
				return nil, snapshotErr
			}

			retVal, err := next.Mutate(ctx, mutation)
			if err != nil {
				return nil, err
			}

			if entityops.EmissionVetoed(ctx) {
				return retVal, err
			}

			topicName := mutation.Type()
			if topicName == "" {
				return retVal, err
			}

			emit := func() {
				if !entityops.InterestedInMutation(galaRuntimes, topicName, op) {
					return
				}

				switch op {
				case entityops.OpUpdate, entityops.OpUpdateOne, entityops.OpDelete, entityops.OpDeleteOne, entityops.OpSoftDelete:
					for _, entityID := range ids {
						payload := entityops.MutationPayload{
							MutationType: topicName,
							Operation:    op,
							EntityID:     entityID,
							ChangeSet:    changeSet,
						}
						payload.OldValues = oldValues[entityID]

						entityops.EmitMutation(ctx, galaRuntimes, payload)
					}
				default:
					if retVal == nil || reflect.TypeOf(retVal).Kind() == reflect.Int {
						return
					}

					entityID, idErr := mutationEventEntityID(retVal)
					if idErr != nil || entityID == "" {
						logx.FromContext(ctx).Error().Err(idErr).Str("mutation_type", topicName).Msg("failed to resolve mutation event id, skipping gala emission")

						return
					}

					entityops.EmitMutation(ctx, galaRuntimes, entityops.MutationPayload{
						MutationType: topicName,
						Operation:    op,
						EntityID:     entityID,
						ChangeSet:    changeSet,
					})
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

// softDeleteClassified reports whether a delete mutation will be rewritten to a soft delete:
// the schema carries the soft-delete column and the context does not skip the rewrite
func softDeleteClassified(ctx context.Context, mutation ent.Mutation) bool {
	if !mutation.Op().Is(ent.OpDelete | ent.OpDeleteOne) {
		return false
	}

	if entx.CheckSkipSoftDelete(ctx) {
		return false
	}

	schema, ok := entityops.LookupSchema(mutation.Type())

	return ok && schema.SoftDeletes()
}

// snapshotMutation captures the mutated row IDs and their pre-mutation values while the
// database still holds the prior rows: updates stash the old values of the changed fields,
// deletes stash the full row so delete listeners can read what was removed. Capture is
// limited to mutations whose topics have at least one interested gala listener
func snapshotMutation(ctx context.Context, mutation ent.Mutation, runtimes []*gala.Gala, op string, changeSet entityops.ChangeSet) ([]string, map[string]map[string]any, error) {
	isDelete := mutation.Op().Is(ent.OpDelete | ent.OpDeleteOne)
	if !isDelete && !mutation.Op().Is(ent.OpUpdate|ent.OpUpdateOne) {
		return nil, nil, nil
	}

	if !entityops.InterestedInMutation(runtimes, mutation.Type(), op) {
		return nil, nil, nil
	}

	mut, ok := mutation.(utils.GenericMutation)
	if !ok {
		return nil, nil, nil
	}

	ids, err := getMutationIDs(ctx, mut)
	if err != nil {
		return nil, nil, err
	}
	if len(ids) == 0 {
		return nil, nil, nil
	}

	schema, ok := entityops.LookupSchema(mutation.Type())
	if !ok || mut.Client() == nil {
		return ids, nil, nil
	}

	changed := changeSet.ChangedFields
	if !isDelete && len(changed) == 0 {
		return ids, nil, nil
	}

	lookupCtx := privacy.DecisionContext(ctx, privacy.Allow)
	oldValues := make(map[string]map[string]any, len(ids))

	for _, id := range ids {
		row, err := schema.Load(lookupCtx, mut.Client(), id)
		if err != nil {
			continue
		}

		fields, err := jsonx.Decode[map[string]any](row)
		if err != nil {
			continue
		}

		if isDelete {
			oldValues[id] = fields

			continue
		}

		oldValues[id] = lo.PickByKeys(fields, changed)
	}

	return ids, oldValues, nil
}
