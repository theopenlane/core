package hooks

import (
	"context"
	"reflect"

	"entgo.io/ent"

	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/events"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/workflowgenerated"
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
				runtimes := resolveGalaRuntimes(galaProviders)
				if len(runtimes) == 0 {
					return
				}

				targets := mutationDispatchTargets(runtimes, mutationDispatchTopics(topicName), op)
				if len(targets) == 0 {
					return
				}

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
		gala.TopicName(schemaType),
		eventqueue.WorkflowMutationTopicName(schemaType),
		eventqueue.NotificationMutationTopicName(schemaType),
	}

	seen := map[gala.TopicName]struct{}{}
	out := make([]gala.TopicName, 0, len(topics))

	for _, topic := range topics {
		if topic == "" {
			continue
		}

		if _, ok := seen[topic]; ok {
			continue
		}

		seen[topic] = struct{}{}
		out = append(out, topic)
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

			if !runtime.Registry().InterestedIn(topic, operation) {
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
