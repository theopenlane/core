package models

import (
	"io"
	"time"
)

// IntegrationHealth records the runtime health state of an installed integration
type IntegrationHealth struct {
	// UnhealthyReason is the user-facing reason the installation stopped syncing
	UnhealthyReason string `json:"unhealthyReason,omitempty"`
	// UnhealthyOperations maps failing operation names to their user-facing reasons
	UnhealthyOperations map[string]string `json:"unhealthyOperations,omitempty"`
	// LastSuccessfulHealthCheck is when the connection health check last passed
	LastSuccessfulHealthCheck *time.Time `json:"lastSuccessfulHealthCheck,omitempty"`
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (h IntegrationHealth) MarshalGQL(w io.Writer) {
	marshalGQLJSON(w, h)
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (h *IntegrationHealth) UnmarshalGQL(v any) error {
	return unmarshalGQLJSON(v, h)
}
