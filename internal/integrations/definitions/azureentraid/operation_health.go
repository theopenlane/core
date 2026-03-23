package azureentraid

import (
	"context"
	"encoding/json"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// HealthCheck holds the result of an Azure Entra ID health check
type HealthCheck struct {
	// Authenticated reports whether the client credentials successfully acquired a token
	Authenticated bool `json:"authenticated"`
}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle() types.OperationHandler {
	return providerkit.WithClient(EntraCredential, h.Run)
}

// Run executes the Azure Entra ID health check by verifying token acquisition
func (HealthCheck) Run(ctx context.Context, cred azcore.TokenCredential) (json.RawMessage, error) {
	_, err := cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{graphScope},
	})
	if err != nil {
		return nil, ErrTokenAcquireFailed
	}

	return providerkit.EncodeResult(HealthCheck{Authenticated: true}, ErrResultEncode)
}
