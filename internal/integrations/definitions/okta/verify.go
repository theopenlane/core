package okta

import (
	"context"
	"fmt"

	oktagosdk "github.com/okta/okta-sdk-golang/v6/okta"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// verify derives the Okta tenant metadata from the org settings
func verify(ctx context.Context, req types.ConnectionRequest[CredentialSchema], c *oktagosdk.APIClient) (InstallationMetadata, error) {
	org, _, err := c.OrgSettingGeneralAPI.GetOrgSettings(ctx).Execute()
	if err != nil {
		return InstallationMetadata{}, fmt.Errorf("%w: %w", ErrOrgSettingsFetchFailed, err)
	}

	if org.GetId() == "" {
		return InstallationMetadata{}, ErrOrgIDMissing
	}

	return InstallationMetadata{
		OrgURL: req.Credential.OrgURL,
		OrgID:  org.GetId(),
	}, nil
}
