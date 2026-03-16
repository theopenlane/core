package operations

import "errors"

var (
	// ErrGalaRequired indicates the gala dependency is missing
	ErrGalaRequired = errors.New("integrationsv2/operations: gala required")
	// ErrInstallationIDRequired indicates the installation identifier is missing
	ErrInstallationIDRequired = errors.New("integrationsv2/operations: installation id required")
	// ErrOperationNameRequired indicates the operation identifier is missing
	ErrOperationNameRequired = errors.New("integrationsv2/operations: operation name required")
	// ErrRunIDRequired indicates the run identifier is missing
	ErrRunIDRequired = errors.New("integrationsv2/operations: run id required")
	// ErrIngestDefinitionNotFound indicates the operation definition could not be resolved for ingest
	ErrIngestDefinitionNotFound = errors.New("integrationsv2/operations: ingest definition not found")
	// ErrIngestSchemaNotFound indicates the generated ingest schema contract was not found
	ErrIngestSchemaNotFound = errors.New("integrationsv2/operations: ingest schema not found")
	// ErrIngestPayloadsInvalid indicates the operation response could not be decoded into ingest payloads
	ErrIngestPayloadsInvalid = errors.New("integrationsv2/operations: ingest payloads invalid")
	// ErrIngestMappingNotFound indicates the definition does not provide a mapping for the emitted payload variant
	ErrIngestMappingNotFound = errors.New("integrationsv2/operations: ingest mapping not found")
	// ErrIngestFilterFailed indicates the CEL filter evaluation failed
	ErrIngestFilterFailed = errors.New("integrationsv2/operations: ingest filter failed")
	// ErrIngestTransformFailed indicates the CEL map evaluation failed
	ErrIngestTransformFailed = errors.New("integrationsv2/operations: ingest transform failed")
	// ErrIngestMappedDocumentInvalid indicates the mapped payload did not satisfy the generated schema contract
	ErrIngestMappedDocumentInvalid = errors.New("integrationsv2/operations: ingest mapped document invalid")
	// ErrIngestRequiredKeyMissing indicates the mapped payload omitted a required generated field
	ErrIngestRequiredKeyMissing = errors.New("integrationsv2/operations: ingest required key missing")
	// ErrIngestUpsertKeyMissing indicates the mapped payload omitted every generated upsert key
	ErrIngestUpsertKeyMissing = errors.New("integrationsv2/operations: ingest upsert key missing")
	// ErrIngestUpsertConflict indicates the generated upsert keys matched more than one record
	ErrIngestUpsertConflict = errors.New("integrationsv2/operations: ingest upsert conflict")
	// ErrIngestUnsupportedSchema indicates the runtime does not yet support the requested generated ingest schema
	ErrIngestUnsupportedSchema = errors.New("integrationsv2/operations: ingest schema unsupported")
	// ErrIngestPersistFailed indicates the mapped record could not be persisted
	ErrIngestPersistFailed = errors.New("integrationsv2/operations: ingest persistence failed")
)
