package azureentraid

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	kiotaauth "github.com/microsoft/kiota-authentication-azure-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// graphScope is the default scope used for Microsoft Graph client credentials requests
const graphScope = "https://graph.microsoft.com/.default"

// tokenCredentialClient builds the Azure client credentials token credential for one installation
func tokenCredentialClient(cfg Config) func(context.Context, types.ConnectionRequest[entraIDCred]) (azcore.TokenCredential, error) {
	return func(ctx context.Context, req types.ConnectionRequest[entraIDCred]) (azcore.TokenCredential, error) {
		if req.Credential.TenantID == "" {
			return nil, ErrCredentialMetadataRequired
		}

		cred, err := azidentity.NewClientSecretCredential(req.Credential.TenantID, cfg.ClientID, cfg.ClientSecret, nil)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("tenant_id", req.Credential.TenantID).Msg("failed to create client secret credential")
			return nil, ErrTokenAcquireFailed
		}

		return cred, nil
	}
}

// graphClient builds the Microsoft Graph service client for one installation
func graphClient(cfg Config) func(context.Context, types.ConnectionRequest[entraIDCred]) (*msgraphsdk.GraphServiceClient, error) {
	return func(_ context.Context, req types.ConnectionRequest[entraIDCred]) (*msgraphsdk.GraphServiceClient, error) {
		if req.Credential.TenantID == "" {
			return nil, ErrCredentialMetadataRequired
		}

		cred, err := azidentity.NewClientSecretCredential(req.Credential.TenantID, cfg.ClientID, cfg.ClientSecret, nil)
		if err != nil {
			return nil, ErrTokenAcquireFailed
		}

		authProvider, err := kiotaauth.NewAzureIdentityAuthenticationProviderWithScopes(cred, []string{graphScope})
		if err != nil {
			return nil, ErrTokenAcquireFailed
		}

		adapter, err := msgraphsdk.NewGraphRequestAdapter(authProvider)
		if err != nil {
			return nil, ErrTokenAcquireFailed
		}

		return msgraphsdk.NewGraphServiceClient(adapter), nil
	}
}
