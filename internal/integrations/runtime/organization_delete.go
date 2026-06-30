package runtime

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/stripe/stripe-go/v84"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/consts"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/internal/ent/generated/orgsubscription"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// SeedOrganizationDeletes starts the durable organization deletion polling loop
// after runtime listeners have been registered.
func (r *Runtime) SeedOrganizationDeletes(ctx context.Context) error {
	if !r.organizationDeleteConfig.Enabled {
		logx.FromContext(ctx).Info().Msg("organization delete listener disabled, skipping seed")
		return nil
	}

	active, err := r.Gala().HasActiveJobForTopic(ctx, operations.OrganizationDeleteTopic)
	if err != nil {
		return err
	}

	if active {
		logx.FromContext(ctx).Debug().Msg("organization delete poller already active, skipping seed")
		return nil
	}

	receipt := r.Gala().EmitWithHeaders(ctx, operations.OrganizationDeleteTopic, operations.OrganizationDeleteEnvelope{}, gala.Headers{
		Tags: []string{"organization-delete"},
	})

	return receipt.Err
}

// HandleOrganizationDeletes clears recovered org deletion markers, then deletes
// overdue non-personal organizations that still have no active or trialing subscription.
func (r *Runtime) HandleOrganizationDeletes(ctx context.Context, _ operations.OrganizationDeleteEnvelope) (int, error) {
	db := r.DB()
	logger := logx.FromContext(ctx)

	cfg := r.organizationDeleteConfig
	if cfg.MaxDeletesPerRun <= 0 {
		cfg.MaxDeletesPerRun = operations.DefaultOrganizationDeleteMaxPerRun
	}

	systemCtx := auth.WithCaller(privacy.DecisionContext(ctx, privacy.Allow), &auth.Caller{
		Capabilities: auth.CapBypassOrgFilter | auth.CapBypassFGA | auth.CapInternalOperation,
	})

	if err := r.clearRecoveredOrganizationDeletions(systemCtx); err != nil {
		return 0, err
	}

	settings, err := db.OrganizationSetting.Query().
		Where(
			organizationsetting.PendingDeletionAtNotNil(),
			organizationsetting.PendingDeletionAtLTE(models.DateTime(time.Now())),
			organizationsetting.HasOrganizationWith(
				organization.IDNEQ(consts.SystemAdminOrgID),
				organization.PersonalOrg(false),
				organization.Not(
					organization.HasOrgSubscriptionsWith(activeOrTrialingSubscriptionPredicates()...),
				),
			),
		).
		WithOrganization().
		Order(organizationsetting.ByUpdatedAt()).
		Limit(cfg.MaxDeletesPerRun).
		All(systemCtx)
	if err != nil {
		logger.Error().Err(err).Msg("failed querying organizations pending deletion")
		return 0, err
	}

	deletedOrgs := make([]string, 0, len(settings))

	for _, setting := range settings {
		org := setting.Edges.Organization
		if org == nil {
			continue
		}

		orgLogger := logger.With().
			Str("organization_id", org.ID).
			Str("setting_id", setting.ID).
			Logger()

		if err := db.Organization.DeleteOneID(org.ID).Exec(systemCtx); err != nil {
			orgLogger.Error().Err(err).Msg("failed to delete organization")
			return len(deletedOrgs), err
		}

		deletedOrgs = append(deletedOrgs, fmt.Sprintf("%s (%s)", org.Name, org.ID))
		orgLogger.Info().Msg("successfully deleted organization")
	}

	logger.Info().
		Int("count", len(deletedOrgs)).
		Str("organizations", strings.Join(deletedOrgs, ", ")).
		Msg("organization deletion summary")

	return len(deletedOrgs), nil
}

func (r *Runtime) clearRecoveredOrganizationDeletions(ctx context.Context) error {
	db := r.DB()
	logger := logx.FromContext(ctx)

	settings, err := db.OrganizationSetting.Query().
		Where(
			organizationsetting.PendingDeletionAtNotNil(),
			organizationsetting.HasOrganizationWith(
				organization.IDNEQ(consts.SystemAdminOrgID),
				organization.PersonalOrg(false),
				organization.HasOrgSubscriptionsWith(activeOrTrialingSubscriptionPredicates()...),
			),
		).
		WithOrganization().
		All(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("failed querying recovered organizations pending deletion")
		return err
	}

	for _, setting := range settings {
		org := setting.Edges.Organization
		if org == nil {
			continue
		}

		if err := db.OrganizationSetting.UpdateOneID(setting.ID).
			ClearPendingDeletionAt().
			Exec(ctx); err != nil {
			logger.Error().Err(err).Str("organization_id", org.ID).Str("setting_id", setting.ID).Msg("failed to clear pending_deletion_at")
			return err
		}

		logger.Info().Str("organization_id", org.ID).Str("setting_id", setting.ID).Msg("cleared pending deletion because billing status recovered")
	}

	return nil
}

func activeOrTrialingSubscriptionPredicates() []predicate.OrgSubscription {
	return []predicate.OrgSubscription{
		orgsubscription.Or(
			orgsubscription.ActiveEQ(true),
			orgsubscription.StripeSubscriptionStatusEQ(string(stripe.SubscriptionStatusTrialing)),
		),
	}
}
