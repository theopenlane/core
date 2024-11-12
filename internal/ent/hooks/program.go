package hooks

import (
	"context"
	"fmt"
	"strings"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/enums"
)

// HookProgramAuthz runs on program mutations to setup or remove relationship tuples
func HookProgramAuthz() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return hook.ProgramFunc(func(ctx context.Context, m *generated.ProgramMutation) (ent.Value, error) {
			// do the mutation, and then create/delete the relationship
			retValue, err := next.Mutate(ctx, m)
			if err != nil {
				// if we error, do not attempt to create the relationships
				return retValue, err
			}

			if m.Op().Is(ent.OpCreate) {
				// create the program member admin and relationship tuple for parent org
				err = programCreateHook(ctx, m)
			} else if m.Op().Is(ent.OpDelete|ent.OpDeleteOne) || entx.CheckIsSoftDelete(ctx) {
				// delete all relationship tuples on delete, or soft delete (Update Op)
				err = programDeleteHook(ctx, m)
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
		}
	}

	objType := strings.ToLower(m.Type())
	org, orgExists := m.OwnerID()

	if exists && orgExists {
		req := fgax.TupleRequest{
			SubjectID:   org,
			SubjectType: "organization",
			ObjectID:    objID,
			ObjectType:  objType,
		}

		log.Debug().Interface("request", req).
			Msg("creating parent relationship tuples")

		orgTuple, err := getTupleKeyFromRole(req, fgax.ParentRelation)
		if err != nil {
			return err
		}

		if _, err := m.Authz.WriteTupleKeys(ctx, []fgax.TupleKey{orgTuple}, nil); err != nil {
			log.Error().Err(err).Msg("failed to create relationship tuple")

			return ErrInternalServerError
		}
	}

	return nil
}

func createProgramMemberAdmin(ctx context.Context, pID string, m *generated.ProgramMutation) error {
	// get userID from context
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("unable to get user id from context, unable to add user to program")

		return err
	}

	// Add user as admin of program
	input := generated.CreateProgramMembershipInput{
		UserID:    userID,
		ProgramID: pID,
		Role:      &enums.RoleAdmin,
	}

	if _, err := m.Client().ProgramMembership.Create().SetInput(input).Save(ctx); err != nil {
		log.Error().Err(err).Msg("error creating program membership for admin")

		return err
	}

	return nil
}

func programDeleteHook(ctx context.Context, m *generated.ProgramMutation) error {
	objID, ok := m.ID()
	if !ok {
		// TODO (sfunk): ensure tuples get cascade deleted
		// continue for now
		return nil
	}

	objType := strings.ToLower(m.Type())
	object := fmt.Sprintf("%s:%s", objType, objID)

	log.Debug().Str("object", object).Msg("deleting relationship tuples")

	if err := m.Authz.DeleteAllObjectRelations(ctx, object); err != nil {
		log.Error().Err(err).Msg("failed to delete relationship tuples")

		return ErrInternalServerError
	}

	log.Debug().Str("object", object).Msg("deleted relationship tuples")

	return nil
}
