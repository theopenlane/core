package graphapi

import (
	"context"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/core/pkg/logx"
)

// UpdateBulkCSVControl is used for bulk updating controls via CSV import
// that includes the ID of the control to update
type UpdateBulkCSVControl struct {
	ID string `json:"id"`
	generated.UpdateControlInput
}

// bulkUpdateControlCSV handles bulk updating controls via CSV import
// TODO (sfunk): pull this to gqlgen-plugins bulk to generate for all schema types
func (r *mutationResolver) bulkUpdateControlCSV(ctx context.Context, inputs []*UpdateBulkCSVControl) (*model.ControlBulkUpdatePayload, error) {
	c := withTransactionalMutation(ctx)
	results := make([]*generated.Control, 0, len(inputs))
	updatedIDs := make([]string, 0, len(inputs))

	// update each control individually to ensure proper validation
	for _, input := range inputs {
		if input.ID == "" {
			logx.FromContext(ctx).Error().Msg("empty id in bulk update for control")
			continue
		}

		// get the existing entity first
		existing, err := c.Control.Get(ctx, input.ID)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("control_id", input.ID).Msg("failed to get control in bulk update operation")
			continue
		}

		// setup update request
		updatedEntity, err := existing.Update().SetInput(input.UpdateControlInput).Save(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("control_id", input.ID).Msg("failed to update control in bulk operation")
			return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionUpdate, Object: "control"})
		}

		results = append(results, updatedEntity)
		updatedIDs = append(updatedIDs, input.ID)
	}

	return &model.ControlBulkUpdatePayload{
		Controls:   results,
		UpdatedIDs: updatedIDs,
	}, nil
}
