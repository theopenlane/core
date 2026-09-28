package email

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// checkHealth validates the email client is configured with a working sender
func checkHealth(_ context.Context, _ types.OperationRequest, client *Client) (json.RawMessage, error) {
	if client.Sender == nil {
		return nil, ErrSenderNotConfigured
	}

	return providerkit.EncodeResult(map[string]any{
		"provider":  client.Config.Provider,
		"fromEmail": client.Config.FromEmail,
	}, nil)
}
