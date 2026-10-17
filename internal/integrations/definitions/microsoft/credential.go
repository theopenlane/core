package microsoft

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

// GraphScope is the default scope used for Microsoft Graph client credentials requests
const GraphScope = "https://graph.microsoft.com/.default"

// NewTokenCredential builds a client credentials token credential for the app in one consented tenant
func NewTokenCredential(cfg AppConfig, tenantID string) (azcore.TokenCredential, error) {
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}

	var opts *azidentity.ClientSecretCredentialOptions
	if cfg.endpoints != nil {
		opts = &azidentity.ClientSecretCredentialOptions{
			ClientOptions:            cfg.clientOptions(),
			DisableInstanceDiscovery: true,
		}
	}

	cred, err := azidentity.NewClientSecretCredential(tenantID, cfg.ApplicationID, cfg.ClientSecret, opts)
	if err != nil {
		return nil, ErrTokenCredentialBuild
	}

	return cred, nil
}
