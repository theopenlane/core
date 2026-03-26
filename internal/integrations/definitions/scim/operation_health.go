package scim

import (
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/providerkit"
)

// healthCheckAckMessage is the static message returned when the SCIM health check is invoked
const healthCheckAckMessage = "scim health check is a no-op for push-based installations"

// HealthCheck is a no-op success result for push-based SCIM installations
type HealthCheck struct {
	// Message describes why the health check does not perform an outbound probe
	Message string `json:"message"`
}

// Run returns the SCIM health check acknowledgement
func (HealthCheck) Run() (json.RawMessage, error) {
	return providerkit.EncodeResult(HealthCheck{
		Message: healthCheckAckMessage,
	}, ErrResultEncode)
}
