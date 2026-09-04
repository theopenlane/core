package system

import (
	"context"
	"encoding/json"
	"time"

	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// Handle adapts the integration lifecycle sweep to the generic operation registration boundary;
// the receiver carries the operator defaults and request config overlays a copy
func (s IntegrationLifecycleSweep) Handle() types.OperationHandler {
	return func(ctx context.Context, req types.OperationRequest) (json.RawMessage, error) {
		sweep := s

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

// Run executes one integration lifecycle sweep and returns the number reaped
func (s IntegrationLifecycleSweep) Run(ctx context.Context, req types.OperationRequest) (int, error) {
	logger := logx.FromContext(ctx)

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
		logger.Error().Err(err).Msg("failed querying expired installations for lifecycle sweep")
		return 0, err
	}

	processed := 0

	for _, id := range ids {
		rowLogger := logger.With().Str("integration_id", id).Logger()

		if s.DryRun {
			rowLogger.Info().Msg("dry run: would reap expired installation")
			processed++

			continue
		}

		reaped, err := req.Services.ReapExpiredInstallation(systemCtx, id)
		if err != nil {
			rowLogger.Error().Err(err).Msg("failed to reap expired installation")

			continue
		}

		if !reaped {
			continue
		}

		rowLogger.Info().Msg("reaped expired installation")
		processed++
	}

	logger.Info().Int("count", processed).Msg("integration lifecycle sweep summary")

	return processed, nil
}
