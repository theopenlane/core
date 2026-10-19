package azureentraid

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/theopenlane/core/v2/internal/integrations/definitions/microsoft"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// tokenCredentialClient builds the Azure client credentials token credential for one installation
func tokenCredentialClient(cfg Config) func(context.Context, types.ConnectionRequest[entraIDCred]) (azcore.TokenCredential, error) {
	return func(ctx context.Context, req types.ConnectionRequest[entraIDCred]) (azcore.TokenCredential, error) {
		return tenantCredential(ctx, cfg, req.Credential)
	}
}

// graphClient builds the Microsoft Graph service client for one installation
func graphClient(cfg Config) func(context.Context, types.ConnectionRequest[entraIDCred]) (*msgraphsdk.GraphServiceClient, error) {
	return func(ctx context.Context, req types.ConnectionRequest[entraIDCred]) (*msgraphsdk.GraphServiceClient, error) {
		cred, err := tenantCredential(ctx, cfg, req.Credential)
		if err != nil {
			return nil, err
		}

		client, err := cfg.ForApplication(req.Credential.ApplicationID, cfg.Legacy).NewGraphClient(cred)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("azureentraid: failed to get graph client")
			return nil, ErrTokenAcquireFailed
		}

		return client, nil
	}
}

// tenantCredential builds the client credentials token credential for the installation's consented tenant
func tenantCredential(ctx context.Context, cfg Config, cred entraIDCred) (azcore.TokenCredential, error) {
	if cred.TenantID == "" {
		return nil, ErrCredentialMetadataRequired
	}

	tokenCred, err := microsoft.NewTokenCredential(cfg.ForApplication(cred.ApplicationID, cfg.Legacy), cred.TenantID)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("tenant_id", cred.TenantID).Msg("azureentraid: failed to create client secret credential")
		return nil, ErrTokenAcquireFailed
	}

	return tokenCred, nil
}
