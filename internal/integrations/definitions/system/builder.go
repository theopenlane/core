package system

import (
	"context"
	"encoding/json"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// sweeper is a scheduled sweep cycle that reports how many records it processed
type sweeper interface {
	Run(ctx context.Context, req types.OperationRequest) (int, error)
}

// Builder returns the system definition hosting the scheduled runtime sweeps
func Builder(paymentReminder PaymentReminderConfig, organizationDelete OrganizationDeleteConfig, integrationLifecycle IntegrationLifecycleConfig) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Family:      "Openlane",
				DisplayName: "Openlane System",
				Description: "Internal scheduled sweeps for organization lifecycle.",
				Category:    "system",
				Active:      true,
				Visible:     false,
			},
			Operations: []types.OperationRegistration{
				PaymentReminderOp.HandlesRequest(sweepHandler(paymentReminder.Sweep())).Registration(DefinitionID, types.OperationRegistration{
					Description:         "Mark canceled organizations for deletion and dispatch deletion notice emails",
					Policy:              types.ExecutionPolicy{Scheduled: true, SkipRunRecord: true},
					Schedule:            &gala.Schedule{MinInterval: PaymentReminderMinInterval, MaxInterval: PaymentReminderMaxInterval},
					CustomerSelectable:  lo.ToPtr(false),
					DisabledForAll:      !paymentReminder.Enabled,
					SkipDefaultLookback: true,
				}),
				OrganizationDeleteOp.HandlesRequest(sweepHandler(organizationDelete.Sweep())).Registration(DefinitionID, types.OperationRegistration{
					Description:         "Delete overdue organizations that still have no active or trialing subscription",
					Policy:              types.ExecutionPolicy{Scheduled: true, SkipRunRecord: true},
					Schedule:            &gala.Schedule{MinInterval: OrganizationDeleteMinInterval, MaxInterval: OrganizationDeleteMaxInterval},
					CustomerSelectable:  lo.ToPtr(false),
					DisabledForAll:      !organizationDelete.Enabled,
					SkipDefaultLookback: true,
				}),
				IntegrationLifecycleOp.HandlesRequest(sweepHandler(integrationLifecycle.Sweep())).Registration(DefinitionID, types.OperationRegistration{
					Description:         "Reap expired integration installations that never connected",
					Policy:              types.ExecutionPolicy{Scheduled: true, SkipRunRecord: true},
					Schedule:            &gala.Schedule{MinInterval: IntegrationLifecycleMinInterval, MaxInterval: IntegrationLifecycleMaxInterval},
					CustomerSelectable:  lo.ToPtr(false),
					DisabledForAll:      !integrationLifecycle.Enabled,
					SkipDefaultLookback: true,
				}),
			},
		}, nil
	})
}

// sweepHandler builds a request handler that runs the sweep and encodes the result
func sweepHandler[S sweeper](defaults S) func(context.Context, types.OperationRequest, S) (json.RawMessage, error) {
	return func(ctx context.Context, req types.OperationRequest, _ S) (json.RawMessage, error) {
		sweep := defaults

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
