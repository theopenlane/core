package usage

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/usage"
	"github.com/theopenlane/core/pkg/enums"
)

// InitializeUsageLimits creates or updates usage rows with the provided limits; NOTE: existing rows will have their limit updated; missing rows will be inserted
func InitializeUsageLimits(ctx context.Context, client *generated.Client, orgID string, limits map[enums.UsageType]int64) error {
	if len(limits) == 0 {
		return nil
	}

	for t, l := range limits {
		u, err := client.Usage.Query().Where(usage.OrganizationID(orgID), usage.ResourceTypeEQ(t)).Only(ctx)

		if generated.IsNotFound(err) {
			_, err = client.Usage.Create().SetOrganizationID(orgID).SetResourceType(t).SetLimit(l).Save(ctx)
		} else if err == nil {
			_, err = client.Usage.UpdateOne(u).SetLimit(l).Save(ctx)
		}
		if err != nil {
			return err
		}

		RecordLimitChange(t, "set")
	}

	return nil
}

// SetUsageLimit sets the usage limit for a specific type on an organization; the row will be created if it does not exist
func SetUsageLimit(ctx context.Context, client *generated.Client, orgID string, t enums.UsageType, limit int64) error {
	return InitializeUsageLimits(ctx, client, orgID, map[enums.UsageType]int64{t: limit})
}

// AddUsageLimit increases the usage limit for an organization by delta; if the row does not exist it will be created with the delta as the limit
func AddUsageLimit(ctx context.Context, client *generated.Client, orgID string, t enums.UsageType, delta int64) error {
	u, err := client.Usage.Query().Where(usage.OrganizationID(orgID), usage.ResourceTypeEQ(t)).Only(ctx)

	if generated.IsNotFound(err) {
		_, err = client.Usage.Create().SetOrganizationID(orgID).SetResourceType(t).SetLimit(delta).Save(ctx)
		if err == nil {
			RecordLimitChange(t, "add")
		}

		return err
	}
	if err != nil {
		return err
	}

	_, err = client.Usage.UpdateOne(u).AddLimit(delta).Save(ctx)
	if err == nil {
		RecordLimitChange(t, "add")
	}

	return err
}

// ClearUsageLimit removes the limit for the provided usage type. A zero limit
// means the organization has unlimited capacity for that resource.
func ClearUsageLimit(ctx context.Context, client *generated.Client, orgID string, t enums.UsageType) error {
	err := SetUsageLimit(ctx, client, orgID, t, 0)
	if err == nil {
		RecordLimitChange(t, "clear")
	}

	return err
}

// IsUnlimited reports whether the usage record has no cap configured
func IsUnlimited(u *generated.Usage) bool {
	return u.Limit == 0
}

// CheckUsageDelta verifies that adding delta to the current usage would not exceed the configured limit
func CheckUsageDelta(ctx context.Context, client *generated.Client, orgID string, t enums.UsageType, delta int64) error {
	if delta <= 0 {
		return nil
	}

	u, err := client.Usage.Query().Where(usage.OrganizationID(orgID), usage.ResourceTypeEQ(t)).Only(ctx)
	if err != nil && !generated.IsNotFound(err) {
		return err
	}

	if err == nil && u.Limit != 0 && u.Used+delta > u.Limit {
		log.Debug().Str("org_id", orgID).Str("type", t.String()).Int64("limit", u.Limit).Int64("used", u.Used).Int64("delta", delta).Msg("usage limit reached")

		return fmt.Errorf("%w for %s", ErrUsageLimitReached, t)
	}

	return nil
}
