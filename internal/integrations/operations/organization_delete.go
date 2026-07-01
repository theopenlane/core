package operations

import (
	"context"
	"time"

	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	// DefaultOrganizationDeleteMaxPerRun is the default maximum number of orgs deleted per cycle
	DefaultOrganizationDeleteMaxPerRun = 25
	// OrganizationDeleteMinInterval is the minimum polling interval for organization deletion
	OrganizationDeleteMinInterval = 24 * time.Hour
	// OrganizationDeleteMaxInterval is the maximum polling interval for organization deletion
	OrganizationDeleteMaxInterval = 24 * time.Hour
)

// OrganizationDeleteConfig contains the configuration for the organization deletion scheduled listener
type OrganizationDeleteConfig struct {
	// MaxDeletesPerRun caps how many overdue organizations are deleted per cycle
	MaxDeletesPerRun int `json:"maxdeletesperrun" koanf:"maxdeletesperrun" jsonschema:"default=25,description=Maximum overdue organizations to delete per run"`
	// Enabled controls whether the organization deletion polling loop is seeded at startup
	Enabled bool `json:"enabled" koanf:"enabled" jsonschema:"description=Whether the organization deletion listener is enabled"`
}

// OrganizationDeleteEnvelope is the durable payload for an organization deletion polling cycle
type OrganizationDeleteEnvelope struct {
	// Schedule is the adaptive scheduling state carried across polling cycles
	Schedule gala.ScheduleState `json:"schedule"`
}

var organizationDeleteSchemaName = jsonx.SchemaID(jsonx.SchemaFrom[OrganizationDeleteEnvelope]())

var (
	// OrganizationDeleteTopic is the Gala topic name for organization deletion polling
	OrganizationDeleteTopic = gala.TopicName("system.organization_delete." + organizationDeleteSchemaName)
	// organizationDeleteListenerName is the Gala listener name for the organization deletion handler
	organizationDeleteListenerName = "system.organization_delete." + organizationDeleteSchemaName + ".handler"
)

// OrganizationDeleteHandler processes one deletion cycle and returns the number of
// orgs deleted, used as the delta for adaptive scheduling
type OrganizationDeleteHandler func(context.Context, OrganizationDeleteEnvelope) (int, error)

// RegisterOrganizationDeleteListener registers the Gala listener for organization deletion polling
func RegisterOrganizationDeleteListener(runtime *gala.Gala, handle OrganizationDeleteHandler, schedule gala.Schedule) error {
	return RegisterScheduledListener(ScheduledListenerConfig[OrganizationDeleteEnvelope]{
		Runtime:  runtime,
		Topic:    OrganizationDeleteTopic,
		Name:     organizationDeleteListenerName,
		Schedule: schedule,
		Handle:   handle,
		State:    func(e OrganizationDeleteEnvelope) gala.ScheduleState { return e.Schedule },
		Wrap: func(_ OrganizationDeleteEnvelope, s gala.ScheduleState) OrganizationDeleteEnvelope {
			return OrganizationDeleteEnvelope{Schedule: s}
		},
	})
}
