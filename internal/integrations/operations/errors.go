package operations

import "errors"

var (
	// ErrGalaRequired indicates the gala dependency is missing
	ErrGalaRequired = errors.New("integrations/operations: gala required")
	// ErrDispatchInputInvalid indicates the queued operation request failed caller-input validation
	ErrDispatchInputInvalid = errors.New("integrations/operations: dispatch input invalid")
	// ErrInstallationIDRequired indicates the installation identifier is missing
	ErrInstallationIDRequired = errors.New("integrations/operations: installation id required")
	// ErrOperationNameRequired indicates the operation identifier is missing
	ErrOperationNameRequired = errors.New("integrations/operations: operation name required")
	// ErrOperationConfigInvalid indicates queued operation config failed caller-input validation
	ErrOperationConfigInvalid = errors.New("integrations/operations: operation config invalid")
	// ErrRunIDRequired indicates the run identifier is missing
	ErrRunIDRequired = errors.New("integrations/operations: run id required")
	// ErrIngestDefinitionNotFound indicates the operation definition could not be resolved for ingest
	ErrIngestDefinitionNotFound = errors.New("integrations/operations: ingest definition not found")
	// ErrIngestSchemaNotFound indicates the generated ingest schema contract was not found
	ErrIngestSchemaNotFound = errors.New("integrations/operations: ingest schema not found")
	// ErrIngestSchemaNotDeclared indicates the payload schema was not declared in the operation's ingest contracts
	ErrIngestSchemaNotDeclared = errors.New("integrations/operations: ingest schema not declared in contracts")
	// ErrIngestMappingNotFound indicates the definition does not provide a mapping for the emitted payload variant
	ErrIngestMappingNotFound = errors.New("integrations/operations: ingest mapping not found")
	// ErrIngestFilterFailed indicates the CEL filter evaluation failed
	ErrIngestFilterFailed = errors.New("integrations/operations: ingest filter failed")
	// ErrIngestInstallationFilterConfigInvalid indicates the installation filter configuration could not be decoded
	ErrIngestInstallationFilterConfigInvalid = errors.New("integrations/operations: ingest installation filter config invalid")
	// ErrIngestTransformFailed indicates the CEL map evaluation failed
	ErrIngestTransformFailed = errors.New("integrations/operations: ingest transform failed")
	// ErrIngestMappedDocumentInvalid indicates the mapped payload did not satisfy the generated schema contract
	ErrIngestMappedDocumentInvalid = errors.New("integrations/operations: ingest mapped document invalid")
	// ErrIngestUpsertKeyMissing indicates the mapped payload omitted every generated upsert key
	ErrIngestUpsertKeyMissing = errors.New("integrations/operations: ingest upsert key missing")
	// ErrIngestUpsertConflict indicates the generated upsert keys matched more than one record
	ErrIngestUpsertConflict = errors.New("integrations/operations: ingest upsert conflict")
	// ErrIngestUnsupportedSchema indicates the runtime does not yet support the requested generated ingest schema
	ErrIngestUnsupportedSchema = errors.New("integrations/operations: ingest schema unsupported")
	// ErrIngestPersistFailed indicates the mapped record could not be persisted
	ErrIngestPersistFailed = errors.New("integrations/operations: ingest persistence failed")
)
