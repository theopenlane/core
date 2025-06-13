package usage

import (
	"context"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/usage"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/events/soiree"
)

// TopicUsageThreshold is emitted when an organization's usage crosses a configured threshold
const TopicUsageThreshold = "usage.threshold"

// thresholdEmitter is the event emitter used for usage threshold notifications
var thresholdEmitter *soiree.EventPool

// thresholds maps resource types to threshold percentages (0-100)
var thresholds = map[enums.UsageType][]int{}

// RegisterThresholdEmitter sets the event emitter used to publish threshold events
func RegisterThresholdEmitter(em *soiree.EventPool) { thresholdEmitter = em }

// RegisterThresholds configures the threshold percentages for a resource type
func RegisterThresholds(t enums.UsageType, percent []int) { thresholds[t] = percent }

// EmitThresholdEvents checks the organization's usage for the given type and emits events for all thresholds that were met or exceeded; this is called from within ent usage hooks
func EmitThresholdEvents(ctx context.Context, client *generated.Client, orgID string, t enums.UsageType) {
	if thresholdEmitter == nil {
		return
	}

	ts := thresholds[t]

	if len(ts) == 0 {
		return
	}

	u, err := client.Usage.Query().Where(usage.OrganizationID(orgID), usage.ResourceTypeEQ(t)).Only(ctx)
	if err != nil {
		return
	}

	if u.Limit == 0 {
		return // unlimited
	}

	percent := int(float64(u.Used) * 100 / float64(u.Limit))

	for _, th := range ts {
		if percent >= th {
			ev := soiree.NewBaseEvent(TopicUsageThreshold, u)

			ev.Properties().Set("org_id", orgID)
			ev.Properties().Set("usage_type", t.String())
			ev.Properties().Set("threshold", th)
			ev.Properties().Set("percent", percent)

			thresholdEmitter.Emit(TopicUsageThreshold, ev)
		}
	}
}
