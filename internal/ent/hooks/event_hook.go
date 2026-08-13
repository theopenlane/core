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
			ctx = entityops.WithEmissionVeto(ctx)

			op := mutation.Op().String()
			if entx.CheckIsSoftDeleteType(ctx, mutation.Type()) {
				op = entityops.OpSoftDelete
			}

			ids, oldValues := snapshotMutation(ctx, mutation, galaRuntimes, op)

			retVal, err := next.Mutate(ctx, mutation)
			if err != nil {
				return nil, err
			}

			if entityops.EmissionVetoed(ctx) {
				return retVal, err
			}

			// a rewritten soft delete emits here on the inner update pass; veto the shared
			// holder so the outer delete pass stays silent
			if op == entityops.OpSoftDelete {
				entityops.VetoEmission(ctx)
			}

			topicName := mutation.Type()
			if topicName == "" {
				return retVal, err
			}

			emit := func() {
				if !entityops.InterestedInMutation(galaRuntimes, topicName, op) {
					return
				}

				changeSet := entityops.ChangeSetFromMutation(mutation)

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

// snapshotMutation captures the mutated row IDs and their pre-mutation values while the
// database still holds the prior rows: updates stash the old values of the changed fields,
// deletes stash the full row so delete listeners can read what was removed. Capture is
// limited to mutations whose topics have at least one interested gala listener
func snapshotMutation(ctx context.Context, mutation ent.Mutation, runtimes []*gala.Gala, op string) ([]string, map[string]map[string]any) {
	isDelete := mutation.Op().Is(ent.OpDelete | ent.OpDeleteOne)
	if !isDelete && !mutation.Op().Is(ent.OpUpdate|ent.OpUpdateOne) {
		return nil, nil
	}

	if !entityops.InterestedInMutation(runtimes, mutation.Type(), op) {
		return nil, nil
	}

	mut, ok := mutation.(utils.GenericMutation)
	if !ok {
		return nil, nil
	}

	ids := getMutationIDs(ctx, mut)
	if len(ids) == 0 {
		return nil, nil
	}

	schema, ok := entityops.LookupSchema(mutation.Type())
	if !ok || mut.Client() == nil {
		return ids, nil
	}

	changed := entityops.ChangeSetFromMutation(mutation).ChangedFields
	if !isDelete && len(changed) == 0 {
		return ids, nil
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

	return ids, oldValues
}
