package ingest

import (
	"encoding/json"

	integrationtypes "github.com/theopenlane/core/common/integrations/types"
)

// DecodeAlertEnvelopes converts an arbitrary value into alert envelopes
func DecodeAlertEnvelopes(value any) ([]integrationtypes.AlertEnvelope, error) {
	switch typed := value.(type) {
	case []integrationtypes.AlertEnvelope:
		return typed, nil
	case integrationtypes.AlertEnvelope:
		return []integrationtypes.AlertEnvelope{typed}, nil
	}

	bytes, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	var out []integrationtypes.AlertEnvelope
	if err := json.Unmarshal(bytes, &out); err != nil {
		return nil, err
	}

	return out, nil
}
