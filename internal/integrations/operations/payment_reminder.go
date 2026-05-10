package operations

import (
	"context"
	"time"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/pkg/gala"
)

const (
	// DefaultPaymentMethodInterval is the default number of days after org creation
	// before an org without a payment method is marked for deletion
	DefaultPaymentMethodInterval = 30
	// DefaultDeletionDays is the default number of days between marking an org
	// for deletion and the actual deletion
	DefaultDeletionDays = 7
	// PaymentReminderMinInterval is the minimum polling interval for payment reminders
	PaymentReminderMinInterval = 6 * time.Hour
	// PaymentReminderMaxInterval is the maximum polling interval for payment reminders
	PaymentReminderMaxInterval = 24 * time.Hour
)

// PaymentReminderConfig contains the configuration for the payment reminder scheduled listener
type PaymentReminderConfig struct {
	// PaymentMethodInterval is the number of days after org creation before
	// an org without a payment method is marked for deletion
	PaymentMethodInterval uint8 `json:"payment_method_interval" koanf:"paymentmethodinterval" jsonschema:"default=30,description=Days after org creation before marking for deletion"`
	// DeletionDays is the number of days between marking an org for deletion
	// and the actual deletion date set on pending_deletion_at
	DeletionDays uint8 `json:"deletion_days" koanf:"deletiondays" jsonschema:"default=7,description=Days between marking and actual deletion"`
	// Enabled controls whether the payment reminder polling loop is seeded at startup
	Enabled bool `json:"enabled" koanf:"enabled" jsonschema:"description=Whether the payment reminder listener is enabled"`
}

// PaymentReminderEnvelope is the durable payload for a payment reminder polling cycle
type PaymentReminderEnvelope struct {
	// Schedule is the adaptive scheduling state carried across polling cycles
	Schedule gala.ScheduleState `json:"schedule"`
}

// paymentReminderSchemaName is the type name derived from the JSON schema reflector
var paymentReminderSchemaName = providerkit.SchemaID(providerkit.SchemaFrom[PaymentReminderEnvelope]())

var (
	// PaymentReminderTopic is the Gala topic name for payment reminder polling
	PaymentReminderTopic = gala.TopicName("system.payment_reminder." + paymentReminderSchemaName)
	// paymentReminderListenerName is the Gala listener name for the payment reminder handler
	paymentReminderListenerName = "system.payment_reminder." + paymentReminderSchemaName + ".handler"
)

// PaymentReminderHandler processes one polling cycle and returns the number of
// orgs notified (used as the delta for adaptive scheduling)
type PaymentReminderHandler func(context.Context, PaymentReminderEnvelope) (int, error)

// RegisterPaymentReminderListener registers the Gala listener for payment reminder polling
func RegisterPaymentReminderListener(runtime *gala.Gala, handle PaymentReminderHandler, schedule gala.Schedule) error {
	return RegisterScheduledListener(ScheduledListenerConfig[PaymentReminderEnvelope]{
		Runtime:  runtime,
		Topic:    PaymentReminderTopic,
		Name:     paymentReminderListenerName,
		Schedule: schedule,
		Handle:   handle,
		State:    func(e PaymentReminderEnvelope) gala.ScheduleState { return e.Schedule },
		Wrap: func(_ PaymentReminderEnvelope, s gala.ScheduleState) PaymentReminderEnvelope {
			return PaymentReminderEnvelope{Schedule: s}
		},
	})
}
