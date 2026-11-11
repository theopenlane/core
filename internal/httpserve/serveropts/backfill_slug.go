package serveropts

import (
	"context"

	"github.com/theopenlane/logx"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/organization"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/v2/pkg/urlx"
)

// slugBackfillBatchSize is the number of rows fetched per page while normalizing slugs
const slugBackfillBatchSize = 50

// backfillSlugs rewrites trust center slugs and organization SSO slugs that pre-date the
// restriction to [a-z0-9-] so stored values match what urlx.Slug derives at create time. Rows
// whose slug is already normalized are left alone, so re-runs are no-ops; a normalized value
// already held by another row gets the same random suffix a fresh create would
func backfillSlugs(ctx context.Context, deps backfillDeps) error {
	if err := backfillTrustCenterSlugs(ctx, deps.Client); err != nil {
		return err
	}

	return backfillOrganizationSlugs(ctx, deps.Client)
}

// backfillTrustCenterSlugs pages every trust center by ID and normalizes slugs that changed
func backfillTrustCenterSlugs(ctx context.Context, client *generated.Client) error {
	var lastID string

	for {
		rows, err := client.TrustCenter.Query().
			Where(trustcenter.IDGT(lastID)).
			Order(trustcenter.ByID()).
			Select(trustcenter.FieldID, trustcenter.FieldSlug).
			Limit(slugBackfillBatchSize).
			All(ctx)
		if err != nil {
			return err
		}

		if len(rows) == 0 {
			logx.FromContext(ctx).Info().Msg("backfill: trust center slugs normalized")

			return nil
		}

		for _, row := range rows {
			normalized := urlx.Slug(row.Slug)
			if normalized == row.Slug {
				continue
			}

			slug, err := urlx.UniqueSlug(normalized, func(candidate string) (bool, error) {
				return client.TrustCenter.Query().
					Where(trustcenter.Slug(candidate), trustcenter.IDNEQ(row.ID)).
					Exist(ctx)
			})
			if err != nil {
				return err
			}

			if err := client.TrustCenter.UpdateOneID(row.ID).SetSlug(slug).Exec(ctx); err != nil {
				return err
			}

			logx.FromContext(ctx).Info().Str("trust_center_id", row.ID).Str("from", row.Slug).Str("to", slug).Msg("backfill: normalized trust center slug")
		}

		lastID = rows[len(rows)-1].ID
	}
}

// backfillOrganizationSlugs pages every organization by ID and normalizes SSO slugs that changed
func backfillOrganizationSlugs(ctx context.Context, client *generated.Client) error {
	var lastID string

	for {
		rows, err := client.Organization.Query().
			Where(organization.IDGT(lastID)).
			Order(organization.ByID()).
			Select(organization.FieldID, organization.FieldSlugName).
			Limit(slugBackfillBatchSize).
			All(ctx)
		if err != nil {
			return err
		}

		if len(rows) == 0 {
			logx.FromContext(ctx).Info().Msg("backfill: organization slugs normalized")

			return nil
		}

		for _, row := range rows {
			normalized := urlx.Slug(row.SlugName)
			if normalized == row.SlugName {
				continue
			}

			slug, err := urlx.UniqueSlug(normalized, func(candidate string) (bool, error) {
				return client.Organization.Query().
					Where(organization.SlugName(candidate), organization.IDNEQ(row.ID)).
					Exist(ctx)
			})
			if err != nil {
				return err
			}

			if err := client.Organization.UpdateOneID(row.ID).SetSlugName(slug).Exec(ctx); err != nil {
				return err
			}

			logx.FromContext(ctx).Info().Str("organization_id", row.ID).Str("from", row.SlugName).Str("to", slug).Msg("backfill: normalized organization slug")
		}

		lastID = rows[len(rows)-1].ID
	}
}
