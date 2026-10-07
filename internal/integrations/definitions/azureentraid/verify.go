package azureentraid

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// verify checks the client credentials can acquire a token and returns the consented tenant identity
func verify(ctx context.Context, req types.ConnectionRequest[entraIDCred], cred azcore.TokenCredential) (InstallationMetadata, error) {
	_, err := cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{graphScope},
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("azure token acquisition failed")

		return InstallationMetadata{}, ErrTokenAcquireFailed
	}

	return InstallationMetadata{TenantID: req.Credential.TenantID}, nil
}
