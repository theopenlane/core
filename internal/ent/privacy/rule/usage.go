package rule

import (
	"context"
	"errors"

	"entgo.io/ent"

	"github.com/rs/zerolog"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/usage"
	"github.com/theopenlane/core/pkg/enums"
)

// AllowIfWithinUsage is a mutation rule that allows a mutation if the organization has not exceeded its usage limit for the given type
func AllowIfWithinUsage(t enums.UsageType) privacy.MutationRuleFunc {
	return privacy.MutationRuleFunc(func(ctx context.Context, m generated.Mutation) error {
		om, ok := m.(interface {
			OwnerID() (string, bool)
			Client() *generated.Client
		})

		if !ok || !m.Op().Is(ent.OpCreate) {
			return privacy.Skip
		}

		orgID, ok := om.OwnerID()
		if !ok || orgID == "" {
			return privacy.Skip
		}

		u, err := om.Client().Usage.Query().Where(usage.OrganizationID(orgID), usage.ResourceTypeEQ(t)).Only(ctx)
		if err != nil && !generated.IsNotFound(err) {
			zerolog.Ctx(ctx).Error().Err(err).Str("org_id", orgID).Str("type", t.String()).Msg("failed to query usage")
			return err
		}

		if err == nil && u.Limit != 0 && u.Used >= u.Limit {
			zerolog.Ctx(ctx).Debug().Str("org_id", orgID).Str("type", t.String()).Int64("limit", u.Limit).Int64("used", u.Used).Msg("usage limit reached")

			return ErrUsageLimitReached
		}

		return privacy.Allow
	})
}

var ErrUsageLimitReached = errors.New("usage limit reached")
