package azureentraid

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	kiotaauth "github.com/microsoft/kiota-authentication-azure-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// graphScope is the default scope used for Microsoft Graph client credentials requests
const graphScope = "https://graph.microsoft.com/.default"

// CredentialClient builds the Azure token credential for one installation
type CredentialClient struct {
	cfg Config
}

// Build constructs the Azure client credentials token credential for one installation
func (c CredentialClient) Build(_ context.Context, req types.ClientBuildRequest) (any, error) {
	meta, err := credentialFromRequest(req)
	if err != nil {
		return nil, err
	}

	cred, err := azidentity.NewClientSecretCredential(meta.TenantID, c.cfg.ClientID, c.cfg.ClientSecret, nil)
	if err != nil {
		return nil, ErrTokenAcquireFailed
	}

	return cred, nil
}

// GraphClient builds the Microsoft Graph service client for one installation
type GraphClient struct {
	cfg Config
}

// Build constructs the Microsoft Graph service client for one installation
func (c GraphClient) Build(_ context.Context, req types.ClientBuildRequest) (any, error) {
	meta, err := credentialFromRequest(req)
	if err != nil {
		return nil, err
	}

	cred, err := azidentity.NewClientSecretCredential(meta.TenantID, c.cfg.ClientID, c.cfg.ClientSecret, nil)
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

// credentialFromRequest decodes and validates the installation credential schema
func credentialFromRequest(req types.ClientBuildRequest) (CredentialSchema, error) {
	var meta CredentialSchema
	if err := jsonx.UnmarshalIfPresent(req.Credential.ProviderData, &meta); err != nil {
		return CredentialSchema{}, ErrMetadataDecode
	}

	if meta.TenantID == "" {
		return CredentialSchema{}, ErrCredentialMetadataRequired
	}

	return meta, nil
}
