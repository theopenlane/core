package providerkit

import (
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// DefaultSampleSize is the standard number of sample items included in operation result details
const DefaultSampleSize = 5

// OperationSuccess builds a successful operation result with structured details
func OperationSuccess(summary string, details any) types.OperationResult {
	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: summary,
		Details: encodeDetails(details),
	}
}

// OperationFailure builds a failed operation result, injecting the error message into the details map
func OperationFailure(summary string, err error, details any) (types.OperationResult, error) {
	if err != nil {
		detailMap, parseErr := jsonx.ToRawMap(details)
		if parseErr != nil {
			detailMap = map[string]json.RawMessage{}
		}

		if _, exists := detailMap["error"]; !exists {
			if errJSON, encErr := jsonx.ToRawMessage(err.Error()); encErr == nil && errJSON != nil {
				detailMap["error"] = errJSON
			}
		}

		details = detailMap
	}

	return types.OperationResult{
		Status:  types.OperationStatusFailed,
		Summary: summary,
		Details: encodeDetails(details),
	}, err
}

// HealthOperation builds a standard health check descriptor for a provider
func HealthOperation(name types.OperationName, description string, client types.ClientName, run types.OperationFunc) types.OperationDescriptor {
	return types.OperationDescriptor{
		Name:        name,
		Kind:        types.OperationKindHealth,
		Description: description,
		Client:      client,
		Run:         run,
	}
}

// encodeDetails safely encodes operation details to JSON, returning nil on failure
func encodeDetails(details any) json.RawMessage {
	raw, err := jsonx.ToRawMessage(details)
	if err != nil {
		return nil
	}

	return raw
}
