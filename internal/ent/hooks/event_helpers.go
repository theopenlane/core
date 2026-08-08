package hooks

import (
	"context"
	"encoding/json"
	"fmt"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// mutationEventID is the identifier structure parsed from mutation return values
type mutationEventID struct {
	ID string `json:"id,omitempty"`
}

// mutationOperation classifies the mutation operation, mapping context-flagged soft
// deletes to the synthetic soft-delete operation
func mutationOperation(ctx context.Context, mutation ent.Mutation) string {
	if gala.HasFlag(ctx, gala.FlagSoftDeleteOperation) && mutation.Op().Is(ent.OpDeleteOne|ent.OpDelete) {
		return gala.SoftDeleteOne
	}

	return mutation.Op().String()
}

// mutationEventEntityID resolves the mutated entity identifier: soft deletes read the
// mutation's own IDs, everything else parses the returned entity value
func mutationEventEntityID(ctx context.Context, mutation ent.Mutation, op string, retVal ent.Value) (string, error) {
	if op == gala.SoftDeleteOne {
		mut, ok := mutation.(utils.GenericMutation)
		if !ok {
			return "", ErrUnableToDetermineEventID
		}

		ids := getMutationIDs(ctx, mut)
		if len(ids) == 0 || ids[0] == "" {
			return "", ErrUnableToDetermineEventID
		}

		if len(ids) > 1 {
			logx.FromContext(ctx).Warn().Strs("mutation_ids", ids).Msg("soft delete mutation returned multiple IDs")
		}

		return ids[0], nil
	}

	out, err := json.Marshal(retVal)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrUnableToDetermineEventID, err)
	}

	event := mutationEventID{}
	if err := json.Unmarshal(out, &event); err != nil {
		return "", fmt.Errorf("%w: %w", ErrUnableToDetermineEventID, err)
	}

	return event.ID, nil
}

// extractChangedEdges flattens the catalog edge deltas into the flat shape consumed by
// pre-commit approval routing
func extractChangedEdges(mutation ent.Mutation) ([]string, map[string][]string, map[string][]string) {
	set := entityops.ChangeSetFromMutation(mutation)

	return set.ChangedEdges, set.AddedIDs, set.RemovedIDs
}
