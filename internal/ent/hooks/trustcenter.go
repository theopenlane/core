package hooks

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"entgo.io/ent"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/iam/fgax"
)

// ErrMissingOrgID is returned when the organization ID is missing from the trust center mutation
var ErrMissingOrgID = fmt.Errorf("missing organization id from trust center mutation")

// HookTrustCenter runs on trust center create mutations
func HookTrustCenter() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterFunc(func(ctx context.Context, m *generated.TrustCenterMutation) (generated.Value, error) {
			orgID, ok := m.OwnerID()
			if !ok {
				return nil, ErrMissingOrgID
			}

			org, err := m.Client().Organization.Query().
				Where(organization.ID(orgID)).
				Select(organization.FieldName).
				Only(ctx)
			if err != nil {
				return nil, err
			}
			// Remove all spaces and non-alphanumeric characters from org.Name, then lowercase
			reg := regexp.MustCompile(`[^a-zA-Z0-9]`)
			cleanedName := reg.ReplaceAllString(org.Name, "")
			slug := strings.ToLower(cleanedName)

			m.SetSlug(slug)

			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			settingID, _ := m.SettingID()
			if settingID != "" {
				log.Info().Msg("trust center setting ID provided, skipping default setting creation")
				return retVal, nil
			}

			trustCenter := retVal.(*generated.TrustCenter)

			// If settings were not created, create default settings
			id, err := GetObjectIDFromEntValue(retVal)
			if err != nil {
				return retVal, err
			}

			setting, err := m.Client().TrustCenterSetting.Create().
				SetTrustCenterID(id).
				SetTitle(fmt.Sprintf("%s Trust Center", org.Name)).
				SetOverview(defaultOverview).
				Save(ctx)
			if err != nil {
				return nil, err
			}

			trustCenter.Edges.Setting = setting

			return trustCenter, nil
		})
	}, ent.OpCreate)
}

const defaultOverview = `
# Welcome to your Trust Center

This is the default overview for your trust center. You can customize this by editing the trust center settings.
`

// HookTrustCenterAuthz runs on trust center mutations to setup or remove relationship tuples
func HookTrustCenterAuthz() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterFunc(func(ctx context.Context, m *generated.TrustCenterMutation) (ent.Value, error) {
			// do the mutation, and then create/delete the relationship
			retValue, err := next.Mutate(ctx, m)
			if err != nil {
				// if we error, do not attempt to create the relationships
				return retValue, err
			}

			if m.Op().Is(ent.OpCreate) {
				// create the trust member admin and relationship tuple for parent org
				err = trustCenterCreateHook(ctx, m)
			} else if isDeleteOp(ctx, m) {
				// delete all relationship tuples on delete, or soft delete (Update Op)
				err = trustCenterDeleteHook(ctx, m)
			}

			return retValue, err
		})
	}
}

// trustCenterCreateHook creates the relationship tuples for the trust center
func trustCenterCreateHook(ctx context.Context, m *generated.TrustCenterMutation) error {
	objID, exists := m.ID()
	org, orgExists := m.OwnerID()

	if exists && orgExists {
		req := fgax.TupleRequest{
			SubjectID:   org,
			SubjectType: generated.TypeOrganization,
			ObjectID:    objID,
			ObjectType:  GetObjectTypeFromEntMutation(m),
		}

		zerolog.Ctx(ctx).Debug().Interface("request", req).
			Msg("creating parent relationship tuples")

		orgTuple, err := getTupleKeyFromRole(req, fgax.ParentRelation)
		if err != nil {
			return err
		}

		if _, err := m.Authz.WriteTupleKeys(ctx, []fgax.TupleKey{orgTuple}, nil); err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("failed to create relationship tuple")

			return ErrInternalServerError
		}
	}

	return nil
}

// trustCenterDeleteHook deletes all relationship tuples for the trust center
func trustCenterDeleteHook(ctx context.Context, m *generated.TrustCenterMutation) error {
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
