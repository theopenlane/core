package operations

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// OperationFailure builds a failed operation result with optional contextual details.
// When err is non-nil it is automatically added under the "error" key unless the
// caller already provided one.
func OperationFailure(summary string, err error, details any) (types.OperationResult, error) {
	if err != nil {
		detailMap := map[string]any{}
		if details != nil {
			if parsed, parseErr := jsonx.ToMap(details); parseErr == nil && parsed != nil {
				detailMap = parsed
			}
		}

		if _, exists := detailMap["error"]; !exists {
			detailMap["error"] = err.Error()
		}

		details = detailMap
	}

	return types.OperationResult{
		Status:  types.OperationStatusFailed,
		Summary: summary,
		Details: encodeOperationDetails(details),
	}, err
}

// OperationSuccess builds a successful operation result.
func OperationSuccess(summary string, details any) types.OperationResult {
	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: summary,
		Details: encodeOperationDetails(details),
	}
}

func encodeOperationDetails(details any) json.RawMessage {
	if details == nil {
		return nil
	}

	var raw json.RawMessage
	if err := jsonx.RoundTrip(details, &raw); err != nil {
		return nil
	}

	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}

	return raw
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

// DefaultSampleSize is the standard number of sample items returned in operation results
const DefaultSampleSize = 5

// TokenType indicates whether to use OAuth or API token extraction
type TokenType int

const (
	// TokenTypeOAuth extracts OAuth access tokens
	TokenTypeOAuth TokenType = iota
	// TokenTypeAPI extracts API tokens
	TokenTypeAPI
)

// HealthCheckRunner creates a health check operation function using the common pattern.
func HealthCheckRunner[T any](tokenType TokenType, endpoint string, failureMsg string, resultFn func(T) (string, any)) types.OperationFunc {
	return func(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
		extract, err := tokenExtractor(tokenType)
		if err != nil {
			return types.OperationResult{}, ErrUnsupportedTokenType
		}

		client, token, err := auth.ClientAndToken(input, extract)
		if err != nil {
			return types.OperationResult{}, err
		}

		var resp T
		if err := auth.GetJSONWithClient(ctx, client, endpoint, token, nil, &resp); err != nil {
			return OperationFailure(failureMsg, err, nil)
		}

		summary, details := resultFn(resp)

		return OperationSuccess(summary, details), nil
	}
}

func tokenExtractor(tokenType TokenType) (auth.TokenExtractor, error) {
	switch tokenType {
	case TokenTypeOAuth:
		return auth.OAuthTokenFromPayload, nil
	case TokenTypeAPI:
		return auth.APITokenFromPayload, nil
	default:
		return nil, ErrUnsupportedTokenType
	}
}
