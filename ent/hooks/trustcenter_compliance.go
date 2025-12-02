package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/hook"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/shared/logx"
)

// HookTrustCenterComplianceAuthz runs on trust center compliance mutations to setup or remove relationship tuples
func HookTrustCenterComplianceAuthz() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterComplianceFunc(func(ctx context.Context, m *generated.TrustCenterComplianceMutation) (ent.Value, error) {
			// do the mutation, and then create/delete the relationship
			retValue, err := next.Mutate(ctx, m)
			if err != nil {
				// if we error, do not attempt to create the relationships
				return retValue, err
			}

			if m.Op().Is(ent.OpCreate) {
				// create the trust member admin and relationship tuple for parent org
				err = trustCenterComplianceCreateHook(ctx, m)
			} else if isDeleteOp(ctx, m) {
				// delete all relationship tuples on delete, or soft delete (Update Op)
				err = trustCenterComplianceDeleteHook(ctx, m)
			}

			return retValue, err
		})
	}
}

func trustCenterComplianceCreateHook(ctx context.Context, m *generated.TrustCenterComplianceMutation) error {
	_, exists := m.ID()

	tcID, tcExists := m.TrustCenterID()
	if exists && tcExists {
		writeTuples := []fgax.TupleKey{}

		standardID, standardExists := m.StandardID()
		if standardExists {
			standardReq := fgax.TupleRequest{
				SubjectID:   tcID,
				SubjectType: "trust_center",
				ObjectID:    standardID,
				ObjectType:  "standard",
				Relation:    "associated_with",
			}
			writeTuples = append(writeTuples, fgax.GetTupleKey(standardReq))
		}

		if _, err := m.Authz.WriteTupleKeys(ctx, writeTuples, nil); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to create relationship tuple")

			return ErrInternalServerError
		}
	}

	return nil
}

func trustCenterComplianceDeleteHook(ctx context.Context, m *generated.TrustCenterComplianceMutation) error {
	trustCenterID, trustCenterExists := m.TrustCenterID()

	standardID, standardExists := m.StandardID()
	if !trustCenterExists || !standardExists {
		return nil
	}

	tupleKey := fgax.GetTupleKey(fgax.TupleRequest{
		SubjectID:   trustCenterID,
		SubjectType: "trust_center",
		ObjectID:    standardID,
		ObjectType:  "standard",
		Relation:    "associated_with",
	})
	if _, err := m.Authz.WriteTupleKeys(ctx, nil, []fgax.TupleKey{tupleKey}); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to delete relationship tuple")

		return ErrInternalServerError
	}

	return nil
}
