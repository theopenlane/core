package integrationgenerated

import "github.com/theopenlane/core/pkg/jsonx"

// DecodeInput converts a mapped field map into the target input type T via JSON round-trip.
// Use this in schema-specific ingest handlers to decode CEL-mapped output into ent create/update inputs:
//
//	input, err := integrationgenerated.DecodeInput[generated.CreateVulnerabilityInput](mapped)
func DecodeInput[T any](data map[string]any) (T, error) {
	var input T
	if err := jsonx.RoundTrip(data, &input); err != nil {
		return input, err
	}

	return input, nil
}
