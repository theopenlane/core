package hooks

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent"
	"github.com/rs/zerolog"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/iam/fgax"
)

var ErrTooManyLogoFiles = errors.New("too many logo files uploaded, only one is allowed")
var ErrTooManyFaviconFiles = errors.New("too many favicon files uploaded, only one is allowed")

func HookTrustCenterSetting() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterSettingFunc(func(ctx context.Context, m *generated.TrustCenterSettingMutation) (generated.Value, error) {
			zerolog.Ctx(ctx).Debug().Msg("trust center setting hook")

			// check for uploaded files (e.g. logo image)
			fileIDs := objects.GetFileIDsFromContext(ctx)
			if len(fileIDs) > 0 {
				var err error

				ctx, err = checkTrustCenterFiles(ctx, m)
				if err != nil {
					return nil, err
				}

				m.AddFileIDs(fileIDs...)
			}
			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

func checkTrustCenterFiles(ctx context.Context, m *generated.TrustCenterSettingMutation) (context.Context, error) {
	logoKey := "logoFile"
	faviconKey := "faviconFile"

	// get the file from the context, if it exists
	logoFile, _ := objects.FilesFromContextWithKey(ctx, logoKey)
	faviconFile, _ := objects.FilesFromContextWithKey(ctx, faviconKey)

	var fileTuples []fgax.TupleKey

	// this should always be true, but check just in case
	if logoFile != nil && logoFile[0].FieldName == logoKey {
		// we should only have one file
		if len(logoFile) > 1 {
			return ctx, ErrTooManyLogoFiles
		}
		m.SetLogoLocalFileID(logoFile[0].ID)

		logoFile[0].Parent.ID, _ = m.ID()
		logoFile[0].Parent.Type = "trust_center_setting"

		ctx = objects.UpdateFileInContextByKey(ctx, logoKey, logoFile[0])

		// add wildcard viewer tuples to allow any user to access the logo file
		wildcardTuples := fgax.CreateWildcardViewerTuple(logoFile[0].ID, generated.TypeFile)
		fileTuples = append(fileTuples, wildcardTuples...)
	}

	if faviconFile != nil && faviconFile[0].FieldName == faviconKey {
		if len(faviconFile) > 1 {
			return ctx, ErrTooManyFaviconFiles
		}

		m.SetFaviconLocalFileID(faviconFile[0].ID)

		faviconFile[0].Parent.ID, _ = m.ID()
		faviconFile[0].Parent.Type = "trust_center_setting"

		ctx = objects.UpdateFileInContextByKey(ctx, faviconKey, faviconFile[0])

		// add wildcard viewer tuples to allow any user to access the favicon file
		wildcardTuples := fgax.CreateWildcardViewerTuple(faviconFile[0].ID, generated.TypeFile)
		fileTuples = append(fileTuples, wildcardTuples...)
	}

	// write the wildcard tuples to allow any user to access the files
	if len(fileTuples) > 0 {
		if _, err := m.Authz.WriteTupleKeys(ctx, fileTuples, nil); err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("failed to create wildcard file access tuples")
			return ctx, fmt.Errorf("failed to create file access permissions: %w", err)
		}
		zerolog.Ctx(ctx).Debug().Interface("tuples", fileTuples).Msg("created wildcard file access tuples")
	}

	return ctx, nil
}

// HookTrustCenterAuthz runs on trust center mutations to setup or remove relationship tuples
func HookTrustCenterSettingAuthz() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterSettingFunc(func(ctx context.Context, m *generated.TrustCenterSettingMutation) (ent.Value, error) {
			// do the mutation, and then create/delete the relationship
			retValue, err := next.Mutate(ctx, m)
			if err != nil {
				// if we error, do not attempt to create the relationships
				return retValue, err
			}

			if m.Op().Is(ent.OpCreate) {
				// create the trust member admin and relationship tuple for parent org
				err = trustCenterSettingCreateHook(ctx, m)
			} else if isDeleteOp(ctx, m) {
				// delete all relationship tuples on delete, or soft delete (Update Op)
				err = trustCenterSettingDeleteHook(ctx, m)
			}

			return retValue, err
		})
	}
}

// trustCenterCreateHook creates the relationship tuples for the trust center
func trustCenterSettingCreateHook(ctx context.Context, m *generated.TrustCenterSettingMutation) error {

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

// trustCenterDeleteHook deletes all relationship tuples for the trust center
func trustCenterSettingDeleteHook(ctx context.Context, m *generated.TrustCenterSettingMutation) error {
	objID, ok := m.ID()
	if !ok {
		return nil
	}

	objType := GetObjectTypeFromEntMutation(m)
	object := fmt.Sprintf("%s:%s", objType, objID)

	zerolog.Ctx(ctx).Debug().Str("object", object).Msg("deleting relationship tuples")

	if err := m.Authz.DeleteAllObjectRelations(ctx, object, userRoles); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("failed to delete relationship tuples")

		return ErrInternalServerError
	}

	zerolog.Ctx(ctx).Debug().Str("object", object).Msg("deleted relationship tuples")

	return nil
}
