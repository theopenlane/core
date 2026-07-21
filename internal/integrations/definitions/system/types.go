package system

import (
	"time"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// DefinitionID is the canonical identifier for the system definition
var DefinitionID = types.NewDefinitionRef("def_01SYSTEM0000000000000000001")

const (
	// DefaultPaymentMethodInterval is the default number of days after org creation
	// before an org without a payment method is marked for deletion
	DefaultPaymentMethodInterval = 30
	// DefaultDeletionDays is the default number of days between marking an org
	// for deletion and the actual deletion
	DefaultDeletionDays = 7
	// PaymentReminderMinInterval is the minimum polling interval for payment reminder sweeps
	PaymentReminderMinInterval = 6 * time.Hour
	// PaymentReminderMaxInterval is the maximum polling interval for payment reminder sweeps
	PaymentReminderMaxInterval = 24 * time.Hour
	// DefaultOrganizationDeleteMaxPerRun is the default maximum number of orgs deleted per sweep
	DefaultOrganizationDeleteMaxPerRun = 25
	// OrganizationDeleteMinInterval is the minimum polling interval for organization deletion sweeps
	OrganizationDeleteMinInterval = 24 * time.Hour
	// OrganizationDeleteMaxInterval is the maximum polling interval for organization deletion sweeps
	OrganizationDeleteMaxInterval = 24 * time.Hour
)

// PaymentReminderConfig contains the operator configuration for the payment reminder sweep
type PaymentReminderConfig struct {
	// PaymentMethodInterval is the number of days after org creation before
	// an org without a payment method is marked for deletion
	PaymentMethodInterval uint8 `json:"paymentmethodinterval" koanf:"paymentmethodinterval" jsonschema:"default=30,description=Days after org creation before marking for deletion"`
	// DeletionDays is the number of days between marking an org for deletion
	// and the actual deletion date set on pending_deletion_at
	DeletionDays uint8 `json:"deletiondays" koanf:"deletiondays" jsonschema:"default=7,description=Days between marking and actual deletion"`
	// Enabled controls whether the payment reminder sweep is seeded at startup
	Enabled bool `json:"enabled" koanf:"enabled" jsonschema:"default=false,description=Whether the payment reminder listener is enabled"`
	// DryRun logs matching organization IDs without mutating state or dispatching emails
	DryRun bool `json:"dryrun" koanf:"dryrun" jsonschema:"default=true,description=If true only log organization IDs that would be processed"`
}

// OrganizationDeleteConfig contains the operator configuration for the organization deletion sweep
type OrganizationDeleteConfig struct {
	// MaxDeletesPerRun caps how many overdue organizations are deleted per sweep
	MaxDeletesPerRun int `json:"maxdeletesperrun" koanf:"maxdeletesperrun" jsonschema:"default=25,description=Maximum overdue organizations to delete per run"`
	// Enabled controls whether the organization deletion sweep is seeded at startup
	Enabled bool `json:"enabled" koanf:"enabled" jsonschema:"description=Whether the organization deletion listener is enabled"`
}

// Sweep maps the operator configuration to its sweep defaults
func (c PaymentReminderConfig) Sweep() PaymentReminderSweep {
	return PaymentReminderSweep{
		PaymentMethodInterval: c.PaymentMethodInterval,
		DeletionDays:          c.DeletionDays,
		DryRun:                c.DryRun,
	}
}

// Sweep maps the operator configuration to its sweep defaults
func (c OrganizationDeleteConfig) Sweep() OrganizationDeleteSweep {
	return OrganizationDeleteSweep{MaxDeletesPerRun: c.MaxDeletesPerRun}
}

// PaymentReminderSweep configures one payment reminder sweep cycle
type PaymentReminderSweep struct {
	// PaymentMethodInterval is the number of days after cancellation before an org is marked for deletion
	PaymentMethodInterval uint8 `json:"paymentMethodInterval,omitempty"`
	// DeletionDays is the number of days between marking an org for deletion and the actual deletion
	DeletionDays uint8 `json:"deletionDays,omitempty"`
	// DryRun logs matching organization IDs without mutating state or dispatching emails
	DryRun bool `json:"dryRun,omitempty"`
}

// OrganizationDeleteSweep configures one organization deletion sweep cycle
type OrganizationDeleteSweep struct {
	// MaxDeletesPerRun caps how many overdue organizations are deleted during the cycle
	MaxDeletesPerRun int `json:"maxDeletesPerRun,omitempty"`
}

var (
	paymentReminderSweepSchema, PaymentReminderOp       = providerkit.OperationSchema[PaymentReminderSweep]()    //nolint:revive
	organizationDeleteSweepSchema, OrganizationDeleteOp = providerkit.OperationSchema[OrganizationDeleteSweep]() //nolint:revive
)
