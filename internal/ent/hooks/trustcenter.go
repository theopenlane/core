package hooks

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"entgo.io/ent"

	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
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

			org, err := m.Client().Organization.Get(ctx, orgID)
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
				SetLogoURL(*org.AvatarRemoteURL).
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
