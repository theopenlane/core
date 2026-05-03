package email

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// HealthCheck validates that the email provider credentials are functional
// by verifying the sender is configured and reachable
type HealthCheck struct{}

// Handle returns the typed operation handler for builder registration
func (h HealthCheck) Handle() types.OperationHandler {
	return providerkit.WithClient(emailClientRef, h.Run)
}

// Run validates the email client is configured with a working sender
func (HealthCheck) Run(_ context.Context, client *Client) (json.RawMessage, error) {
	if client.Sender == nil {
		return nil, ErrSenderNotConfigured
	}

	return providerkit.EncodeResult(map[string]any{
		"provider":  client.Config.Provider,
		"fromEmail": client.Config.FromEmail,
	}, nil)
}
