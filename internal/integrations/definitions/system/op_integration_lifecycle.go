package system

import (
	"context"
	"time"

	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// Run executes one integration lifecycle sweep and returns the number reaped
func (s IntegrationLifecycleSweep) Run(ctx context.Context, req types.OperationRequest) (int, error) {
	if s.MaxPerRun <= 0 {
		s.MaxPerRun = DefaultIntegrationLifecycleMaxPerRun
	}

	systemCtx := systemSweepContext(ctx)

	ids, err := req.DB.Integration.Query().
		Where(integration.ExpiresAtLTE(time.Now())).
		Order(integration.ByExpiresAt()).
		Limit(s.MaxPerRun).
		IDs(systemCtx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed querying expired installations for lifecycle sweep")
		return 0, err
	}

	processed := 0

	for _, id := range ids {
		if s.DryRun {
			logx.FromContext(ctx).Info().Str("integration_id", id).Msg("dry run: would reap expired installation")
			processed++

			continue
		}

		reaped, err := req.Services.ReapExpiredInstallation(systemCtx, id)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("integration_id", id).Msg("failed to reap expired installation")

			continue
		}

		if !reaped {
			continue
		}

		logx.FromContext(ctx).Info().Str("integration_id", id).Msg("reaped expired installation")
		processed++
	}

	logx.FromContext(ctx).Info().Int("count", processed).Msg("integration lifecycle sweep summary")

	return processed, nil
}
