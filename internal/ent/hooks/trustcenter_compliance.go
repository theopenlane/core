package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/iam/fgax"
)

// HookTrustCenterComplianceAuthz runs on trust center compliance mutations to setup or remove relationship tuples
func HookTrustCenterComplianceAuthz() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterComplianceFunc(func(ctx context.Context, m *generated.TrustCenterComplianceMutation) (ent.Value, error) {
			// do the mutation, and then create/delete the relationship
			retValue, err := next.Mutate(ctx, m)
			if err != nil {
				// if we error, do not attempt to create the relationships
				return nil, err
			}

			if m.Op().Is(ent.OpCreate) {
				// create the trust member admin and relationship tuple for parent org
				err = trustCenterComplianceCreateHook(ctx, m)
			} else if isDeleteOp(ctx, m) {
				// delete all relationship tuples on delete, or soft delete (Update Op)
				err = trustCenterComplianceDeleteHook(ctx, m)
			} else if m.Op().Is(ent.OpUpdate | ent.OpUpdateOne) {
				err = trustCenterComplianceUpdateHook(ctx, m)
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

// trustCenterComplianceDeleteHook deletes relationship tuples on trust center compliance deletion
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

// trustCenterComplianceUpdateHook keeps relationship tuples in sync when the standard changes
func trustCenterComplianceUpdateHook(ctx context.Context, m *generated.TrustCenterComplianceMutation) error {
	trustCenterID, trustCenterExists := m.TrustCenterID()
	if !trustCenterExists || trustCenterID == "" {
		if oldTrustCenterID, err := m.OldTrustCenterID(ctx); err == nil && oldTrustCenterID != "" {
			trustCenterID = oldTrustCenterID
		}
	}
	if trustCenterID == "" {
		return nil
	}

	newStandardID, newStandardSet := m.StandardID()
	standardCleared := m.StandardCleared()
	if !newStandardSet && !standardCleared {
		return nil
	}

	oldStandardID, err := m.OldStandardID(ctx)
	if err != nil {
		oldStandardID = ""
	}

	var tuplesToDelete []fgax.TupleKey
	var tuplesToAdd []fgax.TupleKey

	if oldStandardID != "" && (standardCleared || (newStandardSet && newStandardID != oldStandardID)) {
		tupleKey := fgax.GetTupleKey(fgax.TupleRequest{
			SubjectID:   trustCenterID,
			SubjectType: "trust_center",
			ObjectID:    oldStandardID,
			ObjectType:  "standard",
			Relation:    "associated_with",
		})
		tuplesToDelete = append(tuplesToDelete, tupleKey)
	}

	if newStandardSet && newStandardID != "" && newStandardID != oldStandardID {
		tupleKey := fgax.GetTupleKey(fgax.TupleRequest{
			SubjectID:   trustCenterID,
			SubjectType: "trust_center",
			ObjectID:    newStandardID,
			ObjectType:  "standard",
			Relation:    "associated_with",
		})
		tuplesToAdd = append(tuplesToAdd, tupleKey)
	}

	if len(tuplesToAdd) == 0 && len(tuplesToDelete) == 0 {
		return nil
	}

	if _, err := m.Authz.WriteTupleKeys(ctx, tuplesToAdd, tuplesToDelete); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to update relationship tuples")

		return ErrInternalServerError
	}

	return nil
}
