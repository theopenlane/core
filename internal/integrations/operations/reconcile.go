package operations

import (
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
)

// ReconcileEnvelope is the durable payload for one recurring operation cycle, either
// installation-bound (IntegrationID set) or runtime-bound (Runtime true); the type name
// is the durable topic identity and must not change
type ReconcileEnvelope struct {
	gala.OperationContext
	// Schedule is the adaptive scheduling state carried across cycles
	Schedule gala.ScheduleState `json:"schedule"`
}

// ReconcileUniqueKey derives the insert-time uniqueness key for one recurring loop, so any
// emitter of the topic collapses to at most one live loop per installation (or runtime
// definition) and operation
func ReconcileUniqueKey(e ReconcileEnvelope) string {
	src := types.IntegrationSourceFrom(e.OperationContext)

	return gala.IntegrationReconcile.Key(src.IntegrationID, src.DefinitionID, e.Operation)
}

// ReconcileTopic is the durable reconcile topic: the name derives from the envelope type
// under the reconcile namespace, and every emission carries the loop uniqueness key
var ReconcileTopic = gala.NamespacedTopicFor(gala.IntegrationReconcile, gala.WithUniqueKey(ReconcileUniqueKey))
