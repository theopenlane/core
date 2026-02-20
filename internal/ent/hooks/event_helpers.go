package hooks

import (
	"context"
	"encoding/json"
	"fmt"

	"entgo.io/ent"

	"github.com/99designs/gqlgen/graphql"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/internal/ent/eventqueue"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/logx"
)

// EventID represents the identifier structure used in mutation event metadata.
type EventID struct {
	ID string `json:"id,omitempty"`
}

// parseEventID extracts the EventID from the returned mutation value.
func parseEventID(retVal ent.Value) (*EventID, error) {
	out, err := json.Marshal(retVal)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal mutation return value")
		return nil, fmt.Errorf("failed to parse mutation event id: %w", err)
	}

	event := EventID{}
	if err := json.Unmarshal(out, &event); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal mutation return value")
		return nil, err
	}

	return &event, nil
}

// parseSoftDeleteEventID extracts the EventID from a soft-delete mutation.
func parseSoftDeleteEventID(ctx context.Context, mutation ent.Mutation) (*EventID, error) {
	mut, ok := mutation.(utils.GenericMutation)
	if !ok {
		return nil, ErrUnableToDetermineEventID
	}

	ids := getMutationIDs(ctx, mut)
	if len(ids) == 0 || ids[0] == "" {
		return nil, ErrUnableToDetermineEventID
	}

	if len(ids) > 1 {
		logx.FromContext(ctx).Warn().Strs("mutation_ids", ids).Msg("soft delete mutation returned multiple IDs")
	}

	return &EventID{ID: ids[0]}, nil
}

// getOperation determines the operation type from context and mutation data.
func getOperation(ctx context.Context, mutation ent.Mutation) string {
	if graphql.HasOperationContext(ctx) {
		opCtx := graphql.GetOperationContext(ctx)
		if opCtx != nil {
			if opCtx.OperationName == "DeleteOrganization" && mutation.Type() == entgen.TypeOrganization {
				return eventqueue.SoftDeleteOne
			}
		}
	}

	return mutation.Op().String()
}
