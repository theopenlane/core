package hooks

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"entgo.io/ent"

	"github.com/rs/zerolog"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
)

var (
	// compile this only once
	reg = regexp.MustCompile(`[^a-zA-Z0-9]`)
)

// HookTrustCenter runs on trust center create mutations
func HookTrustCenter() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterFunc(func(ctx context.Context, m *generated.TrustCenterMutation) (generated.Value, error) {
			orgID, err := auth.GetOrganizationIDFromContext(ctx)
			if err != nil {
				return nil, err
			}

			exists, err := m.Client().TrustCenter.Query().Exist(ctx)
			if err != nil {
				return nil, err
			}

			if exists {
				return nil, ErrNotSingularTrustCenter
			}

			org, err := m.Client().Organization.Query().
				Where(organization.ID(orgID)).
				Select(organization.FieldName).
				Only(ctx)
			if err != nil {
				return nil, err
			}
			// Remove all spaces and non-alphanumeric characters from org.Name, then lowercase
			cleanedName := reg.ReplaceAllString(org.Name, "")
			slug := strings.ToLower(cleanedName)

			m.SetSlug(slug)

			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			settingID, _ := m.SettingID()
			if settingID != "" {
				zerolog.Ctx(ctx).Debug().Msg("trust center setting ID provided, skipping default setting creation")

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

			wildcardTuples := fgax.CreateWildcardViewerTuple(trustCenter.ID, "trust_center")

			// Create system tuple for system admin access
			systemTuple := fgax.GetTupleKey(fgax.TupleRequest{
				SubjectID:   "openlane_core",
				SubjectType: "system",
				ObjectID:    trustCenter.ID,
				ObjectType:  "trust_center",
				Relation:    "system",
			})

			if _, err := m.Authz.WriteTupleKeys(ctx, append(wildcardTuples, systemTuple), nil); err != nil {
				return ctx, fmt.Errorf("failed to create file access permissions: %w", err)
			}

			return trustCenter, nil
		})
	}, ent.OpCreate)
}

const defaultOverview = `
# Welcome to your Trust Center

This is the default overview for your trust center. You can customize this by editing the trust center settings.
`
