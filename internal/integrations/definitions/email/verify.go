package email

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// verifySender returns the connection verification that checks the provider accepts the configured credentials and reports the delivery configuration under provider
func verifySender[T any](provider string) func(context.Context, types.ConnectionRequest[T], *Client) (InstallationMetadata, error) {
	return func(ctx context.Context, _ types.ConnectionRequest[T], c *Client) (InstallationMetadata, error) {
		if c.Sender == nil {
			return InstallationMetadata{}, ErrSenderNotConfigured
		}

		if err := c.Sender.Verify(ctx); err != nil {
			return InstallationMetadata{}, fmt.Errorf("%w: %w", ErrSenderVerifyFailed, err)
		}

		return InstallationMetadata{
			Provider:  provider,
			FromEmail: c.Config.FromEmail,
		}, nil
	}
}
