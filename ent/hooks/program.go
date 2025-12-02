package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/shared/enums"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/hook"
	"github.com/theopenlane/ent/generated/privacy"
	"github.com/theopenlane/shared/logx"
)

// HookProgramAuthz runs on program mutations to setup or remove relationship tuples
// and prevents updates to archived programs - except if the update contains status changes too
func HookProgramAuthz() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return hook.ProgramFunc(func(ctx context.Context, m *generated.ProgramMutation) (ent.Value, error) {
			if m.Op().Is(ent.OpUpdate|ent.OpUpdateOne) && !isDeleteOp(ctx, m) {
				if err := checkArchivedProgram(ctx, m); err != nil {
					return nil, err
				}
			}

			// do the mutation, and then create/delete the relationship
			retValue, err := next.Mutate(ctx, m)
			if err != nil {
				// if we error, do not attempt to create the relationships
				return retValue, err
			}

			if m.Op().Is(ent.OpCreate) {
				// create the program member admin and relationship tuple for parent org
				err = programCreateHook(ctx, m)
			}

			return retValue, err
		})
	}
}

func programCreateHook(ctx context.Context, m *generated.ProgramMutation) error {
	objID, exists := m.ID()
	if exists {
		// create the admin program member if not using an API token (which is not associated with a user)
		if !auth.IsAPITokenAuthentication(ctx) {
			if err := createProgramMemberAdmin(ctx, objID, m); err != nil {
				return err
			}
		} else {
			if err := addTokenEditPermissions(ctx, m, objID, GetObjectTypeFromEntMutation(m)); err != nil {
				return err
			}
		}
	}

	org, orgExists := m.OwnerID()
	if exists && orgExists {
		req := fgax.TupleRequest{
			SubjectID:   org,
			SubjectType: generated.TypeOrganization,
			ObjectID:    objID,
			ObjectType:  GetObjectTypeFromEntMutation(m),
		}

		logx.FromContext(ctx).Debug().Interface("request", req).
			Msg("creating parent relationship tuples")

		orgTuple, err := getTupleKeyFromRole(req, fgax.ParentRelation)
		if err != nil {
			return err
		}

		if _, err := m.Authz.WriteTupleKeys(ctx, []fgax.TupleKey{orgTuple}, nil); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to create relationship tuple")

			return ErrInternalServerError
		}
	}

	return nil
}

func createProgramMemberAdmin(ctx context.Context, pID string, m *generated.ProgramMutation) error {
	// get userID from context
	userID, err := auth.GetSubjectIDFromContext(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("unable to get user id from context, unable to add user to program")

		return err
	}

	// Add user as admin of program
	input := generated.CreateProgramMembershipInput{
		UserID:    userID,
		ProgramID: pID,
		Role:      &enums.RoleAdmin,
	}

	// allow before the permissions have been added for the program itself
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	if err := m.Client().ProgramMembership.Create().SetInput(input).Exec(allowCtx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error creating program membership for admin")

		return err
	}

	return nil
}

// checkArchivedProgram prevents updates to archived programs if need be
func checkArchivedProgram(ctx context.Context, m *generated.ProgramMutation) error {
	id, _ := m.ID()

	program, err := m.Client().Program.Get(ctx, id)
	if err != nil {
		return err
	}

	if program.Status != enums.ProgramStatusArchived {
		return nil
	}

	status, exists := m.Status()
	if !exists {
		return ErrArchivedProgramUpdateNotAllowed
	}

	if status == enums.ProgramStatusArchived {
		return ErrArchivedProgramUpdateNotAllowed
	}

	return nil
}
