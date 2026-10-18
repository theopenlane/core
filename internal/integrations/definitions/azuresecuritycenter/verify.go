package azuresecuritycenter

import (
	"context"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// verify checks the app can authenticate and list subscriptions; zero subscriptions passes so the role can be assigned after consent
func verify(ctx context.Context, req types.ConnectionRequest[securityCenterCred], c *Client) (InstallationMetadata, error) {
	if _, err := c.arm.ListSubscriptionIDs(ctx, req.Credential.TenantID); err != nil {
		return InstallationMetadata{}, err
	}

	return InstallationMetadata{TenantID: req.Credential.TenantID}, nil
}

// verifyLighthouse checks at least one subscription in the tenant is delegated with the installation's connection code
func verifyLighthouse(ctx context.Context, req types.ConnectionRequest[lighthouseCred], c *Client) (InstallationMetadata, error) {
	if _, err := c.accessibleSubscriptions(ctx); err != nil {
		return InstallationMetadata{}, err
	}

	return InstallationMetadata{TenantID: req.Credential.TenantID}, nil
}
