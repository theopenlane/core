package azuresecuritycenter

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

// staticAzureCredential adapts a pre-obtained OAuth bearer token to azcore.TokenCredential.
type staticAzureCredential struct {
	token string
}

// GetToken satisfies azcore.TokenCredential for a pre-obtained bearer token.
func (s staticAzureCredential) GetToken(_ context.Context, _ policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: s.token}, nil
}
