package operations

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// OperationFailure builds a failed operation result with optional contextual details
func OperationFailure(summary string, err error, details any) (types.OperationResult, error) {
	if err != nil {
		detailMap, parseErr := jsonx.ToRawMap(details)
		if parseErr != nil {
			detailMap = map[string]json.RawMessage{}
		}

		if _, exists := detailMap["error"]; !exists {
			errJSON, _ := jsonx.ToRawMessage(err.Error())
			if errJSON != nil {
				detailMap["error"] = errJSON
			}
		}

		details = detailMap
	}

	return types.OperationResult{
		Status:  types.OperationStatusFailed,
		Summary: summary,
		Details: encodeOperationDetails(details),
	}, err
}

// OperationSuccess builds a successful operation result
func OperationSuccess(summary string, details any) types.OperationResult {
	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: summary,
		Details: encodeOperationDetails(details),
	}
}

// encodeOperationDetails safely encodes operation details to JSON
func encodeOperationDetails(details any) json.RawMessage {
	raw, err := jsonx.ToRawMessage(details)
	if err != nil {
		return nil
	}

	return raw
}

// HealthOperation builds a standard health check descriptor
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

// HealthCheckRunner creates a health check operation function using the common shared generic pattern
func HealthCheckRunner[T any](extractor auth.TokenExtractor, endpoint string, failureMsg string, resultFn func(T) (string, any)) types.OperationFunc {
	return func(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
		client, err := auth.ResolveAuthenticatedClient(input, extractor, "", nil)
		if err != nil {
			return types.OperationResult{}, err
		}

		var resp T
		if err := client.GetJSON(ctx, endpoint, &resp); err != nil {
			return OperationFailure(failureMsg, err, nil)
		}

		summary, details := resultFn(resp)

		return OperationSuccess(summary, details), nil
	}
}
