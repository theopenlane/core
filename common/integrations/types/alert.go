package types

import "encoding/json"

// AlertEnvelope captures raw provider alert payloads for downstream mapping.
type AlertEnvelope struct {
	// AlertType is the provider alert type
	AlertType string `json:"alert_type"`
	// Resource is the upstream resource identifier
	Resource string `json:"resource,omitempty"`
	// Payload is the raw alert payload
	Payload json.RawMessage `json:"payload"`
}
