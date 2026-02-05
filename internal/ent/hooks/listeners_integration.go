package hooks

import (
	"github.com/theopenlane/core/internal/integrations/runner"
	"github.com/theopenlane/core/pkg/events/soiree"
)

// RegisterIntegrationOperationListeners registers listeners for integration operation events
func RegisterIntegrationOperationListeners(eventer *Eventer) {
	if eventer == nil {
		return
	}

	eventer.AddListenerBinding(soiree.BindListener(
		soiree.IntegrationOperationRequestedTopic,
		func(ctx *soiree.EventContext, payload soiree.IntegrationOperationRequestedPayload) error {
			return runner.HandleIntegrationOperationRequested(ctx, payload)
		},
	))
}
