package types

import "encoding/json"

// AlertEnvelope captures raw provider alert payloads for downstream mapping.
type AlertEnvelope struct {
	AlertType string          `json:"alert_type"`
	Resource  string          `json:"resource,omitempty"`
	Payload   json.RawMessage `json:"payload"`
}
