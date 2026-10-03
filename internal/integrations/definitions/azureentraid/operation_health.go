package azureentraid

import (
	"context"
	"encoding/json"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// HealthCheck holds the result of an Azure Entra ID health check
type HealthCheck struct {
	// Authenticated reports whether the client credentials successfully acquired a token
	Authenticated bool `json:"authenticated"`
}

// checkHealth executes the Azure Entra ID health check by verifying token acquisition
func checkHealth(ctx context.Context, _ types.OperationRequest, cred azcore.TokenCredential) (json.RawMessage, error) {
	_, err := cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{graphScope},
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("azure token acquisition failed")
		return nil, ErrTokenAcquireFailed
	}

	return providerkit.EncodeResult(HealthCheck{Authenticated: true}, ErrResultEncode)
}
