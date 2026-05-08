package operations

import (
	"context"
	"time"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/pkg/gala"
)

// RecurringCampaignEnvelope is the durable payload for a recurring campaign polling cycle
type RecurringCampaignEnvelope struct {
	// Schedule is the adaptive scheduling state carried across polling cycles
	Schedule gala.ScheduleState `json:"schedule"`
}

// recurringCampaignSchemaName is the type name derived from the JSON schema reflector
var recurringCampaignSchemaName = providerkit.SchemaID(providerkit.SchemaFrom[RecurringCampaignEnvelope]())

var (
	// RecurringCampaignTopic is the Gala topic name for recurring campaign polling
	RecurringCampaignTopic = gala.TopicName("campaign.recurring." + recurringCampaignSchemaName)
	// recurringCampaignListenerName is the Gala listener name for the recurring campaign handler
	recurringCampaignListenerName = "campaign.recurring." + recurringCampaignSchemaName + ".handler"
)

// RecurringCampaignHandler processes one polling cycle and returns the number of
// campaigns dispatched (used as the delta for adaptive scheduling)
type RecurringCampaignHandler func(context.Context, RecurringCampaignEnvelope) (int, error)

// RegisterRecurringCampaignListener registers the Gala listener for recurring campaign polling
func RegisterRecurringCampaignListener(runtime *gala.Gala, handle RecurringCampaignHandler, schedule gala.Schedule) error {
	return RegisterScheduledListener(ScheduledListenerConfig[RecurringCampaignEnvelope]{
		Runtime:  runtime,
		Topic:    RecurringCampaignTopic,
		Name:     recurringCampaignListenerName,
		Schedule: schedule,
		Handle:   handle,
		State:    func(e RecurringCampaignEnvelope) gala.ScheduleState { return e.Schedule },
		Wrap: func(_ RecurringCampaignEnvelope, s gala.ScheduleState) RecurringCampaignEnvelope {
			return RecurringCampaignEnvelope{Schedule: s}
		},
	})
}

// NextCampaignRunAt computes the next run time from the given base time using
// calendar-based frequency and interval arithmetic. All frequencies are
// calendar-relative (month boundaries, not fixed durations) so time.AddDate is
// used rather than time.Add
func NextCampaignRunAt(from time.Time, frequency enums.Frequency, interval int, timezone string) time.Time {
	loc := time.UTC
	if timezone != "" {
		if parsed, err := time.LoadLocation(timezone); err == nil {
			loc = parsed
		}
	}

	base := from.In(loc)

	switch frequency {
	case enums.FrequencyMonthly:
		return base.AddDate(0, interval, 0).In(time.UTC)
	case enums.FrequencyQuarterly:
		return base.AddDate(0, quarterMonths*interval, 0).In(time.UTC)
	case enums.FrequencyBiAnnually:
		return base.AddDate(0, biannualMonths*interval, 0).In(time.UTC)
	case enums.FrequencyYearly:
		return base.AddDate(interval, 0, 0).In(time.UTC)
	default:
		return from
	}
}

const (
	// quarterMonths is the number of months in a quarter
	quarterMonths = 3
	// biannualMonths is the number of months in a half year
	biannualMonths = 6
)
