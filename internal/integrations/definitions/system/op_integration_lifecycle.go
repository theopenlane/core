package system

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/theopenlane/core/common/enums"
	generated "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// abandonedPendingAge is how long a pending installation must sit untouched before it is
// considered abandoned; live auth state expires after minutes, so anything beyond this window
// can only be completed by restarting the flow
const abandonedPendingAge = 168 * time.Hour

// LifecycleAction identifies the runtime verb a matched rule dispatches to
type LifecycleAction string

const (
	// LifecycleActionRemove deletes the matched installation and its credentials locally
	LifecycleActionRemove LifecycleAction = "remove"
	// LifecycleActionProbe re-verifies an errored installation and clears it on success
	LifecycleActionProbe LifecycleAction = "probe"
	// LifecycleActionMark flags the matched installation unhealthy
	LifecycleActionMark LifecycleAction = "mark"
)

// LifecycleRule declares one match criterion and the action dispatched when it matches
type LifecycleRule struct {
	// Name identifies the rule in logs
	Name string
	// Match reports whether the rule applies to the installation row at the given time
	Match func(row *generated.Integration, now time.Time) bool
	// Action is the runtime verb dispatched on match
	Action LifecycleAction
	// Reason is the user-facing reason recorded when Action is mark
	Reason string
}

// integrationLifecycleRules declares the sweep criteria evaluated in order with
// first-match-wins semantics
var integrationLifecycleRules = []LifecycleRule{
	{
		Name: "reap-abandoned-pending",
		Match: func(row *generated.Integration, now time.Time) bool {
			return row.Status == enums.IntegrationStatusPending && row.UpdatedAt.Before(now.Add(-abandonedPendingAge))
		},
		Action: LifecycleActionRemove,
	},
	{
		Name: "finalize-deleted",
		Match: func(row *generated.Integration, _ time.Time) bool {
			return row.Status == enums.IntegrationStatusDeleted
		},
		Action: LifecycleActionRemove,
	},
	{
		Name: "reprobe-errored",
		Match: func(row *generated.Integration, _ time.Time) bool {
			return row.Status == enums.IntegrationStatusErrored
		},
		Action: LifecycleActionProbe,
	},
}

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

// Run executes one integration lifecycle sweep and returns the number of processed installations
func (s IntegrationLifecycleSweep) Run(ctx context.Context, req types.OperationRequest) (int, error) {
	logger := logx.FromContext(ctx)

	if s.MaxPerRun <= 0 {
		s.MaxPerRun = DefaultIntegrationLifecycleMaxPerRun
	}

	systemCtx := systemSweepContext(ctx)

	rows, err := req.DB.Integration.Query().
		Where(integration.StatusIn(enums.IntegrationStatusPending, enums.IntegrationStatusDeleted, enums.IntegrationStatusErrored)).
		Order(integration.ByUpdatedAt()).
		Limit(s.MaxPerRun).
		All(systemCtx)
	if err != nil {
		logger.Error().Err(err).Msg("failed querying integrations for lifecycle sweep")
		return 0, err
	}

	processed := s.sweepRows(systemCtx, req.Services, rows)

	logger.Info().Int("count", processed).Msg("integration lifecycle sweep summary")

	return processed, nil
}

// sweepRows evaluates the lifecycle rules against each row and dispatches the first matching
// rule's action, returning the number of processed installations
func (s IntegrationLifecycleSweep) sweepRows(ctx context.Context, services types.RuntimeServices, rows []*generated.Integration) int {
	logger := logx.FromContext(ctx)
	now := time.Now()
	processed := 0

	for _, row := range rows {
		rule, matched := matchLifecycleRule(row, now)
		if !matched {
			continue
		}

		rowLogger := logger.With().Str("integration_id", row.ID).Str("rule", rule.Name).Str("action", string(rule.Action)).Logger()

		if s.DryRun {
			rowLogger.Info().Msg("dry run: would dispatch integration lifecycle action")
			processed++

			continue
		}

		if err := dispatchLifecycleAction(ctx, services, row, rule); err != nil {
			switch rule.Action {
			case LifecycleActionProbe:
				rowLogger.Info().Err(err).Msg("integration recovery probe failed, installation stays errored")
				processed++
			default:
				rowLogger.Error().Err(err).Msg("failed to dispatch integration lifecycle action")
			}

			continue
		}

		rowLogger.Info().Msg("dispatched integration lifecycle action")
		processed++
	}

	return processed
}

// matchLifecycleRule returns the first declared rule matching the installation row
func matchLifecycleRule(row *generated.Integration, now time.Time) (LifecycleRule, bool) {
	for _, rule := range integrationLifecycleRules {
		if rule.Match(row, now) {
			return rule, true
		}
	}

	return LifecycleRule{}, false
}

// dispatchLifecycleAction routes one matched rule to its runtime verb
func dispatchLifecycleAction(ctx context.Context, services types.RuntimeServices, row *generated.Integration, rule LifecycleRule) error {
	switch rule.Action {
	case LifecycleActionRemove:
		return services.RemoveInstallation(ctx, row.ID)
	case LifecycleActionProbe:
		return services.ProbeIntegrationRecovery(ctx, row)
	case LifecycleActionMark:
		return services.MarkIntegrationUnhealthy(ctx, row, rule.Reason)
	default:
		return fmt.Errorf("%w: %s", ErrLifecycleActionUnknown, rule.Action)
	}
}
