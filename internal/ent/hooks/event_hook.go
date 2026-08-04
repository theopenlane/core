package hooks

import (
	"context"
	"reflect"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/mutations"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/entx"
)

// EmitGalaEventHook returns a hook that emits Gala mutation envelopes after mutations.
func EmitGalaEventHook(galaProviders ...func() *gala.Gala) ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, mutation ent.Mutation) (ent.Value, error) {
			if entx.CheckIsSoftDeleteType(ctx, mutation.Type()) {
				return next.Mutate(ctx, mutation)
			}

			ctx = workflows.WithSkipEventEmission(ctx)

			oldValues := snapshotOldValues(ctx, mutation, galaProviders)

			retVal, err := next.Mutate(ctx, mutation)
			if err != nil {
				return nil, err
			}

			if workflows.ShouldSkipEventEmission(ctx) {
				return retVal, err
			}

			op := getOperation(ctx, mutation)

			if op != eventqueue.SoftDeleteOne && retVal != nil && reflect.TypeOf(retVal).Kind() == reflect.Int {
				return retVal, err
			}

			topicName := mutation.Type()
			if topicName == "" {
				return retVal, err
			}

			emit := func() {
				runtimes := resolveGalaRuntimes(galaProviders)
				if len(runtimes) == 0 {
					return
				}

				targets := mutationDispatchTargets(runtimes, mutationDispatchTopics(topicName), op)
				if len(targets) == 0 {
					return
				}

				var (
					eventID  *EventID
					parseErr error
				)

				if op == eventqueue.SoftDeleteOne {
					eventID, parseErr = parseSoftDeleteEventID(ctx, mutation)
					if parseErr != nil {
						logx.FromContext(ctx).Info().Err(parseErr).Msg("failed to parse event ID for soft delete, skipping gala emission")

						return
					}
				} else {
					eventID, parseErr = parseEventID(retVal)
					if parseErr != nil {
						logx.FromContext(ctx).Error().Err(parseErr).Msg("failed to parse event ID, skipping gala emission")

						return
					}
				}

				if eventID == nil || eventID.ID == "" {
					logx.FromContext(ctx).Error().Msg("event ID is nil or empty, skipping gala emission")

					return
				}

				payload := newMutationPayloadForDispatch(mutation, op, eventID.ID, oldValues)
				metadata := eventqueue.NewMutationGalaMetadata(eventID.ID, payload)

				for _, target := range targets {
					if galaErr := enqueueGalaMutation(ctx, target.runtime, string(target.topic), payload, metadata); galaErr != nil {
						logx.FromContext(ctx).Error().Err(galaErr).Str("topic", string(target.topic)).Msg("gala mutation dispatch failed")
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

// extractChangedEdges flattens the entityops edge-change catalog into the mutation payload shape:
// the changed edge names plus added/removed IDs keyed by edge
func extractChangedEdges(mutation ent.Mutation) ([]string, map[string][]string, map[string][]string) {
	changes := entityops.ExtractChangedEdges(mutation)
	if len(changes) == 0 {
		return nil, nil, nil
	}

	edges := make([]string, 0, len(changes))
	added := map[string][]string{}
	removed := map[string][]string{}

	for _, change := range changes {
		edges = append(edges, change.Edge)

		if len(change.AddedIDs) > 0 {
			added[change.Edge] = change.AddedIDs
		}

		if change.RemovedIDs != nil {
			removed[change.Edge] = change.RemovedIDs
		}
	}

	if len(added) == 0 {
		added = nil
	}

	if len(removed) == 0 {
		removed = nil
	}

	return edges, added, removed
}

func resolveGalaRuntimes(providers []func() *gala.Gala) []*gala.Gala {
	if len(providers) == 0 {
		return nil
	}

	seen := map[*gala.Gala]struct{}{}
	runtimes := make([]*gala.Gala, 0, len(providers))

	for _, provider := range providers {
		if provider == nil {
			continue
		}

		runtime := provider()
		if runtime == nil {
			continue
		}

		if _, ok := seen[runtime]; ok {
			continue
		}

		seen[runtime] = struct{}{}
		runtimes = append(runtimes, runtime)
	}

	if len(runtimes) == 0 {
		return nil
	}

	return runtimes
}

func mutationDispatchTopics(schemaType string) []gala.TopicName {
	topics := []gala.TopicName{
		eventqueue.MutationTopicName(eventqueue.MutationConcernDirect, schemaType),
		eventqueue.MutationTopicName(eventqueue.MutationConcernWorkflow, schemaType),
		eventqueue.MutationTopicName(eventqueue.MutationConcernNotification, schemaType),
	}

	out := make([]gala.TopicName, 0, len(topics))
	for _, topic := range topics {
		if topic != "" {
			out = append(out, topic)
		}
	}

	return out
}

type mutationDispatchTarget struct {
	runtime *gala.Gala
	topic   gala.TopicName
}

func mutationDispatchTargets(runtimes []*gala.Gala, topics []gala.TopicName, operation string) []mutationDispatchTarget {
	if len(runtimes) == 0 || len(topics) == 0 {
		return nil
	}

	seen := map[mutationDispatchTarget]struct{}{}
	targets := make([]mutationDispatchTarget, 0, len(runtimes)*len(topics))

	for _, runtime := range runtimes {
		if runtime == nil {
			continue
		}

		for _, topic := range topics {
			if topic == "" {
				continue
			}

			if !runtime.InterestedIn(topic, operation) {
				continue
			}

			target := mutationDispatchTarget{runtime: runtime, topic: topic}
			if _, ok := seen[target]; ok {
				continue
			}

			seen[target] = struct{}{}
			targets = append(targets, target)
		}
	}

	if len(targets) == 0 {
		return nil
	}

	return targets
}

// newMutationPayloadForDispatch builds shared mutation payload metadata for asynchronous dispatch hooks.
func newMutationPayloadForDispatch(mutation ent.Mutation, operation, entityID string, oldValues map[string]any) eventqueue.MutationGalaPayload {
	changedFields, clearedFields := mutations.ChangedAndClearedFields(mutation)
	changedEdges, addedIDs, removedIDs := extractChangedEdges(mutation)
	proposedChanges := mutations.BuildProposedChanges(mutation, changedFields)

	return eventqueue.MutationGalaPayload{
		MutationType:    mutation.Type(),
		Operation:       operation,
		EntityID:        entityID,
		ChangedFields:   changedFields,
		ClearedFields:   clearedFields,
		ChangedEdges:    changedEdges,
		AddedIDs:        addedIDs,
		RemovedIDs:      removedIDs,
		ProposedChanges: proposedChanges,
		OldValues:       oldValues,
	}
}

// snapshotOldValues captures pre-update field values before the mutation applies, while the
// database still holds the prior row. Capture is limited to single-row updates whose topics
// have at least one interested gala listener, so unobserved mutations cost nothing; ent
// caches the loaded old row on the mutation, so hooks that already called OldField share it
func snapshotOldValues(ctx context.Context, mutation ent.Mutation, galaProviders []func() *gala.Gala) map[string]any {
	source, ok := mutation.(mutations.OldValueSource)
	if !ok || !mutation.Op().Is(ent.OpUpdateOne) {
		return nil
	}

	runtimes := resolveGalaRuntimes(galaProviders)
	if len(runtimes) == 0 {
		return nil
	}

	if len(mutationDispatchTargets(runtimes, mutationDispatchTopics(mutation.Type()), mutation.Op().String())) == 0 {
		return nil
	}

	changedFields, _ := mutations.ChangedAndClearedFields(mutation)

	return mutations.BuildOldValues(ctx, source, changedFields)
}
