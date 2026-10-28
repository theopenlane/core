package serveropts

import (
	"context"

	"github.com/samber/lo"
	"github.com/theopenlane/logx"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/organization"
)

func backfillOrgBrandingSettings(ctx context.Context, deps backfillDeps) error {

	var batchSize = 50

	var lastKnownID string

	for {
		orgs, err := deps.Client.Organization.Query().
			Order(organization.ByID()).
			Where(organization.IDGT(lastKnownID)).
			Limit(batchSize).
			All(ctx)

		if err != nil {

			logx.FromContext(ctx).Error().Err(err).
				Str("filtered_from", lastKnownID).
				Msg("backfill: failed to query orgs")

			return err
		}

		if len(orgs) == 0 {
			break
		}

		for _, org := range orgs {

			input := generated.CreateOrganizationBrandingSettingInput{
				OrganizationID: lo.ToPtr(org.ID),
			}

			err := deps.Client.OrganizationBrandingSetting.Create().
				SetInput(input).
				Exec(ctx)
			if err != nil {
				logx.FromContext(ctx).Err(err).
					Str("organization_id", org.ID).
					Msg("could not create branding settings for org")

				return err
			}

			lastKnownID = org.ID
		}
	}

	logx.FromContext(ctx).Info().Msg("bacfkill: created org branding settings for all orgs")
	return nil
}
