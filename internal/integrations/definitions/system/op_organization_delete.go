package system

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/consts"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// Handle adapts the organization deletion sweep to the generic operation registration boundary;
// the receiver carries the operator defaults and request config overlays a copy
func (o OrganizationDeleteSweep) Handle() types.OperationHandler {
	return func(ctx context.Context, req types.OperationRequest) (json.RawMessage, error) {
		sweep := o

		if err := jsonx.UnmarshalIfPresent(req.Config, &sweep); err != nil {
			return nil, ErrOperationConfigInvalid
		}

		processed, err := sweep.Run(ctx, req)
		if err != nil {
			return nil, err
		}

		return providerkit.EncodeResult(types.ScheduledCycleResult{Processed: processed}, ErrResultEncode)
	}
}

// Run executes one organization deletion sweep and returns the number of deleted organizations
func (sweep OrganizationDeleteSweep) Run(ctx context.Context, req types.OperationRequest) (int, error) {
	db := req.DB
	logger := logx.FromContext(ctx)

	if sweep.MaxDeletesPerRun <= 0 {
		sweep.MaxDeletesPerRun = DefaultOrganizationDeleteMaxPerRun
	}

	systemCtx := systemSweepContext(ctx)

	if err := clearRecoveredOrganizationDeletions(systemCtx, req); err != nil {
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
		Limit(sweep.MaxDeletesPerRun).
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
		Msg("organization deletion summary")

	return len(deletedOrgs), nil
}

// clearRecoveredOrganizationDeletions clears pending deletion markers on organizations whose
// billing status recovered since being marked
func clearRecoveredOrganizationDeletions(ctx context.Context, req types.OperationRequest) error {
	db := req.DB
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
