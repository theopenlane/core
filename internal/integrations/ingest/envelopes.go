package ingest

import (
	integrationtypes "github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// DecodeAlertEnvelopes converts an arbitrary value into alert envelopes
func DecodeAlertEnvelopes(value any) ([]integrationtypes.AlertEnvelope, error) {
	switch typed := value.(type) {
	case []integrationtypes.AlertEnvelope:
		return typed, nil
	case integrationtypes.AlertEnvelope:
		return []integrationtypes.AlertEnvelope{typed}, nil
	}

	var out []integrationtypes.AlertEnvelope
	if err := jsonx.RoundTrip(value, &out); err != nil {
		return nil, err
	}

	return out, nil
}
