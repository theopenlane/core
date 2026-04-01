package integrationgenerated

import (
	"encoding/json"

	"github.com/theopenlane/core/pkg/jsonx"
)

// DecodeInput converts a mapped JSON document into the target input type T via JSON round-trip.
// Use this in schema-specific ingest handlers to decode CEL-mapped output into ent create/update inputs:
//
//	input, err := integrationgenerated.DecodeInput[generated.CreateVulnerabilityInput](mappedRaw)
func DecodeInput[T any](data json.RawMessage) (T, error) {
	var input T
	if err := jsonx.RoundTrip(data, &input); err != nil {
		return input, err
	}

	return input, nil
}
