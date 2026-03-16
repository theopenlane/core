package scim

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const healthCheckAckMessage = "scim health check is a no-op for push-based installations"

// HealthCheck is a no-op success result for SCIM credential validation
type HealthCheck struct {
	// Message describes why the health check does not perform an outbound probe
	Message string `json:"message"`
}

// Handle adapts the SCIM health check to the generic operation registration boundary
func (h HealthCheck) Handle() types.OperationHandler {
	return func(_ context.Context, _ types.OperationRequest) (json.RawMessage, error) {
		return h.Run()
	}
}

// Run returns the SCIM health check acknowledgement
func (HealthCheck) Run() (json.RawMessage, error) {
	return providerkit.EncodeResult(HealthCheck{
		Message: healthCheckAckMessage,
	}, ErrResultEncode)
}
