package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/rs/zerolog"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/iam/fgax"
)

// HookTrustCenterControlAuthz runs on trust center control mutations to setup or remove relationship tuples
func HookTrustCenterControlAuthz() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterControlFunc(func(ctx context.Context, m *generated.TrustCenterControlMutation) (ent.Value, error) {
			// do the mutation, and then create/delete the relationship
			retValue, err := next.Mutate(ctx, m)
			if err != nil {
				// if we error, do not attempt to create the relationships
				return retValue, err
			}

			if m.Op().Is(ent.OpCreate) {
				// create the trust center control association tuple
				err = trustCenterControlCreateHook(ctx, m)
			} else if isDeleteOp(ctx, m) {
				// delete the trust center control association tuple on delete, or soft delete (Update Op)
				err = trustCenterControlDeleteHook(ctx, m)
			}

			return retValue, err
		})
	}
}

func trustCenterControlCreateHook(ctx context.Context, m *generated.TrustCenterControlMutation) error {
	trustCenterControlID, exists := m.ID()
	controlID, controlExists := m.ControlID()
	if !exists || !controlExists {
		return nil
	}

	// Create the trust_center_association tuple that links the control to the trust_center_control
	// This allows the control to be publicly viewable via the FGA model
	trustCenterAssociationReq := fgax.TupleRequest{
		SubjectID:   trustCenterControlID,
		SubjectType: "trust_center_control",
		ObjectID:    controlID,
		ObjectType:  "control",
		Relation:    "trust_center_association",
	}

	writeTuples := []fgax.TupleKey{fgax.GetTupleKey(trustCenterAssociationReq)}

	if _, err := m.Authz.WriteTupleKeys(ctx, writeTuples, nil); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("failed to create trust center control relationship tuple")
		return ErrInternalServerError
	}

	zerolog.Ctx(ctx).Debug().
		Str("trust_center_control_id", trustCenterControlID).
		Str("control_id", controlID).
		Msg("created trust center control association tuple")

	return nil
}

func trustCenterControlDeleteHook(ctx context.Context, m *generated.TrustCenterControlMutation) error {
	trustCenterControlID, exists := m.ID()
	controlID, controlExists := m.ControlID()
	if !exists || !controlExists {
		return nil
	}

	// Delete the trust_center_association tuple that links the control to the trust_center_control
	trustCenterAssociationReq := fgax.TupleRequest{
		SubjectID:   trustCenterControlID,
		SubjectType: "trust_center_control",
		ObjectID:    controlID,
		ObjectType:  "control",
		Relation:    "trust_center_association",
	}

	deleteTuples := []fgax.TupleKey{fgax.GetTupleKey(trustCenterAssociationReq)}

	if _, err := m.Authz.WriteTupleKeys(ctx, nil, deleteTuples); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("failed to delete trust center control relationship tuple")
		return ErrInternalServerError
	}

	zerolog.Ctx(ctx).Debug().
		Str("trust_center_control_id", trustCenterControlID).
		Str("control_id", controlID).
		Msg("deleted trust center control association tuple")

	return nil
}
