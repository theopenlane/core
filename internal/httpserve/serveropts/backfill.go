package serveropts

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/stoewer/go-strcase"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

// backfillBypassCaps lets the backfill write organizations and memberships without a request caller while
// skipping the org-filter, FGA, and managed-group guards the membership hooks would otherwise apply
const backfillBypassCaps = auth.CapBypassOrgFilter | auth.CapBypassFGA | auth.CapInternalOperation | auth.CapBypassManagedGroup

// WithBackfill runs one-time, idempotent startup backfills for fields introduced by recent migrations:
// organization slug names and the SSO exemption on existing organization owners. It is gated by the
// Backfill.Enabled config flag and runs in the background so it never blocks server startup
func WithBackfill(ctx context.Context, dbClient *ent.Client) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if dbClient == nil || !s.Config.Settings.Backfill.Enabled {
			return
		}

		go func() {
			backfillCtx := privacy.DecisionContext(ctx, privacy.Allow)
			backfillCtx = auth.WithCaller(backfillCtx, &auth.Caller{Capabilities: backfillBypassCaps})

			backfillOrganizationSlugs(backfillCtx, dbClient)
			backfillOwnerSSOExemptions(backfillCtx, dbClient)
		}()
	})
}

// backfillOrganizationSlugs derives slug_name from the organization name for organizations that pre-date
// the field, matching the kebab-case format applied to newly created organizations
func backfillOrganizationSlugs(ctx context.Context, dbClient *ent.Client) {
	orgs, err := dbClient.Organization.Query().
		Where(organization.Or(organization.SlugNameIsNil(), organization.SlugNameEQ(""))).
		All(ctx)
	if err != nil {
		log.Error().Err(err).Msg("backfill: failed to query organizations missing a slug name")
		return
	}

	updated := 0

	for _, org := range orgs {
		if err := dbClient.Organization.UpdateOneID(org.ID).
			SetSlugName(strcase.KebabCase(org.Name)).
			Exec(ctx); err != nil {
			log.Error().Err(err).Str("organization_id", org.ID).Msg("backfill: failed to set organization slug name")

			continue
		}

		updated++
	}

	log.Info().Int("updated", updated).Int("candidates", len(orgs)).Msg("backfill: organization slug names populated")
}

// backfillOwnerSSOExemptions marks existing organization owners as SSO exempt so the exemption is explicit
// on the membership record, matching how owners are seeded for newly created organizations
func backfillOwnerSSOExemptions(ctx context.Context, dbClient *ent.Client) {
	updated, err := dbClient.OrgMembership.Update().
		Where(orgmembership.RoleEQ(enums.RoleOwner), orgmembership.SSOExempt(false)).
		SetSSOExempt(true).
		SetSSOExemptReason("organization owner").
		Save(ctx)
	if err != nil {
		log.Error().Err(err).Msg("backfill: failed to set owner SSO exemptions")
		return
	}

	log.Info().Int("updated", updated).Msg("backfill: owner SSO exemptions populated")
}
