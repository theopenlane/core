package email

import (
	"context"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// verify validates the email client is configured with a working sender and reports its delivery configuration
func verify(_ context.Context, _ types.ConnectionRequest[Credential], c *Client) (InstallationMetadata, error) {
	if c.Sender == nil {
		return InstallationMetadata{}, ErrSenderNotConfigured
	}

	return InstallationMetadata{
		Provider:  c.Config.Provider,
		FromEmail: c.Config.FromEmail,
	}, nil
}
