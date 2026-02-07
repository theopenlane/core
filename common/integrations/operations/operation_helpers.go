package operations

import "github.com/theopenlane/core/common/integrations/types"

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
