package operations

import (
	"context"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
)

// OperationFailure builds a failed operation result with an error detail
func OperationFailure(summary string, err error) types.OperationResult {
	if err == nil {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: summary,
		}
	}

	return types.OperationResult{
		Status:  types.OperationStatusFailed,
		Summary: summary,
		Details: map[string]any{"error": err.Error()},
	}
}

// OperationSuccess builds a successful operation result.
func OperationSuccess(summary string, details map[string]any) types.OperationResult {
	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: summary,
		Details: details,
	}
}

// HealthOperation builds a standard health check descriptor.
func HealthOperation(name types.OperationName, description string, client types.ClientName, run types.OperationFunc) types.OperationDescriptor {
	return types.OperationDescriptor{
		Name:        name,
		Kind:        types.OperationKindHealth,
		Description: description,
		Client:      client,
		Run:         run,
	}
}

// TokenType indicates whether to use OAuth or API token extraction
type TokenType int

const (
	// TokenTypeOAuth extracts OAuth access tokens
	TokenTypeOAuth TokenType = iota
	// TokenTypeAPI extracts API tokens
	TokenTypeAPI
)

// HealthCheckRunner creates a health check operation function using the common pattern.
func HealthCheckRunner[T any](tokenType TokenType, endpoint string, failureMsg string, resultFn func(T) (string, map[string]any)) types.OperationFunc {
	return func(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
		var client *auth.AuthenticatedClient
		var token string
		var err error

		switch tokenType {
		case TokenTypeOAuth:
			client, token, err = auth.ClientAndOAuthToken(input)
		case TokenTypeAPI:
			client, token, err = auth.ClientAndAPIToken(input)
		}

		if err != nil {
			return types.OperationResult{}, err
		}

		var resp T
		if err := auth.GetJSONWithClient(ctx, client, endpoint, token, nil, &resp); err != nil {
			return OperationFailure(failureMsg, err), err
		}

		summary, details := resultFn(resp)

		return OperationSuccess(summary, details), nil
	}
}
