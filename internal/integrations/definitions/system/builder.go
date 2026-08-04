package system

import (
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

// Builder returns the system definition hosting the scheduled runtime sweeps; it exposes
// no credentials, clients, or connections and is never visible in catalog surfaces
func Builder(paymentReminder PaymentReminderConfig, organizationDelete OrganizationDeleteConfig) registry.Builder {
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
				{
					Name:                PaymentReminderOp.Name(),
					Description:         "Mark canceled organizations for deletion and dispatch deletion notice emails",
					Topic:               DefinitionID.OperationTopic(PaymentReminderOp.Name()),
					ConfigSchema:        paymentReminderSweepSchema,
					Policy:              types.ExecutionPolicy{Scheduled: true, SkipRunRecord: true},
					Schedule:            lo.ToPtr(gala.NewSchedule(gala.WithMinInterval(PaymentReminderMinInterval), gala.WithMaxInterval(PaymentReminderMaxInterval))),
					Handle:              paymentReminder.Sweep().Handle(),
					CustomerSelectable:  lo.ToPtr(false),
					DisabledForAll:      !paymentReminder.Enabled,
					SkipDefaultLookback: true,
				},
				{
					Name:                OrganizationDeleteOp.Name(),
					Description:         "Delete overdue organizations that still have no active or trialing subscription",
					Topic:               DefinitionID.OperationTopic(OrganizationDeleteOp.Name()),
					ConfigSchema:        organizationDeleteSweepSchema,
					Policy:              types.ExecutionPolicy{Scheduled: true, SkipRunRecord: true},
					Schedule:            lo.ToPtr(gala.NewSchedule(gala.WithMinInterval(OrganizationDeleteMinInterval), gala.WithMaxInterval(OrganizationDeleteMaxInterval))),
					Handle:              organizationDelete.Sweep().Handle(),
					CustomerSelectable:  lo.ToPtr(false),
					DisabledForAll:      !organizationDelete.Enabled,
					SkipDefaultLookback: true,
				},
			},
		}, nil
	})
}
