package ingest

import integrationtypes "github.com/theopenlane/core/common/integrations/types"

// operationBinding associates an ingest function and pre-execution requirements with an operation name
type operationBinding struct {
	// Ingest is the function called after a successful operation to materialize result envelopes
	Ingest IngestFunc
	// EnsurePayloads indicates that include_payloads must be forced true before execution
	EnsurePayloads bool
}

// operationBindings maps well-known operation names to their ingest bindings.
// Add an entry here when introducing a new operation that produces ingest-able envelopes.
var operationBindings = map[integrationtypes.OperationName]operationBinding{
	integrationtypes.OperationVulnerabilitiesCollect: {
		Ingest:         VulnerabilityIngestFunc(),
		EnsurePayloads: true,
	},
	integrationtypes.OperationDirectorySync: {
		Ingest: DirectoryAccountIngestFunc(),
	},
}

// BindingForOperation returns the ingest function and pre-execution requirements for the given
// operation name, if one is registered. The second return value indicates whether include_payloads
// must be forced true before execution. The third return value reports whether any binding exists.
func BindingForOperation(name integrationtypes.OperationName) (IngestFunc, bool, bool) {
	binding, ok := operationBindings[name]
	return binding.Ingest, binding.EnsurePayloads, ok
}
