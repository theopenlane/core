package hooks

import (
	"context"
	"fmt"

	"entgo.io/ent"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/iam/fgax"
)

func HookTrustCenterDoc() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterDocFunc(func(ctx context.Context, m *generated.TrustCenterDocMutation) (generated.Value, error) {
			zerolog.Ctx(ctx).Debug().Msg("trust center doc hook")

			// check for uploaded files (e.g. logo image)
			fileIDs := objects.GetFileIDsFromContext(ctx)
			if len(fileIDs) > 0 {
				var err error

				ctx, err = checkTrustCenterDocFile(ctx, m)
				if err != nil {
					return nil, err
				}
			}
			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}

func HookTrustCenterDocAuthz() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterDocFunc(func(ctx context.Context, m *generated.TrustCenterDocMutation) (ent.Value, error) {
			// do the mutation, and then create/delete the relationship
			retValue, err := next.Mutate(ctx, m)
			if err != nil {
				// if we error, do not attempt to create the relationships
				return retValue, err
			}
			if m.Op().Is(ent.OpCreate) {
				// create the trust member admin and relationship tuple for parent org
				err = trustCenterDocCreateHook(ctx, m)
			} else if isDeleteOp(ctx, m) {
				// delete all relationship tuples on delete, or soft delete (Update Op)
				err = trustCenterDocDeleteHook(ctx, m)
			}

			return retValue, err
		})
	}
}

func trustCenterDocCreateHook(ctx context.Context, m *generated.TrustCenterDocMutation) error {
	objID, exists := m.ID()
	tcID, tcExists := m.TrustCenterID()

	if exists && tcExists {
		req := fgax.TupleRequest{
			SubjectID:   tcID,
			SubjectType: "trust_center",
			ObjectID:    objID,
			ObjectType:  GetObjectTypeFromEntMutation(m),
		}

		zerolog.Ctx(ctx).Debug().Interface("request", req).
			Msg("creating parent relationship tuples")

		orgTuple, err := getTupleKeyFromRole(req, fgax.ParentRelation)
		if err != nil {
			return err
		}

		zerolog.Ctx(ctx).Debug().Msg(fmt.Sprintf("org tuple: %+v, request: %+v", orgTuple, req))

		if _, err := m.Authz.WriteTupleKeys(ctx, []fgax.TupleKey{orgTuple}, nil); err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("failed to create relationship tuple")

			return ErrInternalServerError
		}
	}

	return nil
}

func trustCenterDocDeleteHook(ctx context.Context, m *generated.TrustCenterDocMutation) error {
	objID, ok := m.ID()
	if !ok {
		return nil
	}

	objType := GetObjectTypeFromEntMutation(m)
	object := fmt.Sprintf("%s:%s", objType, objID)

	zerolog.Ctx(ctx).Debug().Str("object", object).Msg("deleting relationship tuples")

	if err := m.Authz.DeleteAllObjectRelations(ctx, object, userRoles); err != nil {
		log.Error().Err(err).Msg("failed to delete relationship tuples")

		return ErrInternalServerError
	}

	zerolog.Ctx(ctx).Debug().Str("object", object).Msg("deleted relationship tuples")

	return nil
}

func checkTrustCenterDocFile(ctx context.Context, m *generated.TrustCenterDocMutation) (context.Context, error) {
	dockey := "trustCenterDocFile"

	// get the file from the context, if it exists
	docFile, _ := objects.FilesFromContextWithKey(ctx, dockey)
	if docFile == nil {
		return ctx, nil
	}

	// this should always be true, but check just in case
	if docFile[0].FieldName == dockey {
		// we should only have one file
		if len(docFile) > 1 {
			return ctx, ErrNotSingularUpload
		}
		m.SetFileID(docFile[0].ID)

		docFile[0].Parent.ID, _ = m.ID()
		docFile[0].Parent.Type = "trust_center_doc"

		ctx = objects.UpdateFileInContextByKey(ctx, dockey, docFile[0])
	}

	return ctx, nil
}
