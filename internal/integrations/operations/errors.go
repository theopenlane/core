package operations

import "errors"

var (
	// ErrGalaRequired indicates the gala dependency is missing
	ErrGalaRequired = errors.New("integrations/operations: gala required")
	// ErrDispatchInputInvalid indicates the queued operation request failed caller-input validation
	ErrDispatchInputInvalid = errors.New("integrations/operations: dispatch input invalid")
	// ErrInstallationIDRequired indicates the installation identifier is missing
	ErrInstallationIDRequired = errors.New("integrations/operations: installation id required")
	// ErrIntegrationIDRequired indicates resolution requires an explicit integration ID
	ErrIntegrationIDRequired = errors.New("integrations/operations: integration id required")
	// ErrInstallationDefinitionMismatch indicates the installation doesn't match the definition
	ErrInstallationDefinitionMismatch = errors.New("integrations/operations: installation definition mismatch")
	// ErrOperationConfigInvalid indicates queued operation config failed caller-input validation
	ErrOperationConfigInvalid = errors.New("integrations/operations: operation config invalid")
	// ErrRunIDRequired indicates the run identifier is missing
	ErrRunIDRequired = errors.New("integrations/operations: run id required")
	// ErrIngestDefinitionNotFound indicates the operation definition could not be resolved for ingest
	ErrIngestDefinitionNotFound = errors.New("integrations/operations: ingest definition not found")
	// ErrIngestSchemaNotFound indicates the generated ingest schema contract was not found
	ErrIngestSchemaNotFound = errors.New("integrations/operations: ingest schema not found")
	// ErrIngestSchemaNotDeclared indicates the schema isn't in the operation's ingest contracts
	ErrIngestSchemaNotDeclared = errors.New("integrations/operations: ingest schema not declared in contracts")
	// ErrIngestMappingNotFound indicates no mapping exists for the emitted payload variant
	ErrIngestMappingNotFound = errors.New("integrations/operations: ingest mapping not found")
	// ErrIngestFilterFailed indicates the CEL filter evaluation failed
	ErrIngestFilterFailed = errors.New("integrations/operations: ingest filter failed")
	// ErrIngestInstallationFilterConfigInvalid indicates the filter config couldn't decode
	ErrIngestInstallationFilterConfigInvalid = errors.New("integrations/operations: ingest installation filter config invalid")
	// ErrIngestTransformFailed indicates the CEL map evaluation failed
	ErrIngestTransformFailed = errors.New("integrations/operations: ingest transform failed")
	// ErrIngestMappedDocumentInvalid indicates the mapped payload failed the generated schema contract
	ErrIngestMappedDocumentInvalid = errors.New("integrations/operations: ingest mapped document invalid")
	// ErrIngestUpsertKeyMissing indicates the mapped payload omitted every generated upsert key
	ErrIngestUpsertKeyMissing = errors.New("integrations/operations: ingest upsert key missing")
	// ErrIngestUpsertConflict indicates the generated upsert keys matched more than one record
	ErrIngestUpsertConflict = errors.New("integrations/operations: ingest upsert conflict")
	// ErrIngestUnsupportedSchema indicates the runtime doesn't support the requested ingest schema
	ErrIngestUnsupportedSchema = errors.New("integrations/operations: ingest schema unsupported")
	// ErrIngestPersistFailed indicates the mapped record could not be persisted
	ErrIngestPersistFailed = errors.New("integrations/operations: ingest persistence failed")
	// ErrIngestRecordExcluded indicates the record was skipped after an earlier run marked it failing
	ErrIngestRecordExcluded = errors.New("integrations/operations: ingest record excluded")
	// ErrIngestIntegrationUnresolved indicates the integration couldn't be resolved for ingest
	ErrIngestIntegrationUnresolved = errors.New("integrations/operations: ingest integration unresolved")
	// ErrIngestInstanceIDRequired indicates the installation has no source instance id for provenance
	ErrIngestInstanceIDRequired = errors.New("integrations/operations: ingest installation instance id required")

	// ErrIngestIntegrationRemoved indicates the installation was removed while ingest jobs were queued
	ErrIngestIntegrationRemoved = errors.New("integrations/operations: ingest integration removed")
	// ErrOperationDisabled indicates the operation is disabled and the reconcile cycle should stop
	ErrOperationDisabled = errors.New("integrations/operations: operation disabled")
	// ErrExportFailed indicates the Drive file export request failed
	ErrExportFailed = errors.New("integrations/operations: file export failed")
	// ErrResultEncode indicates an operation result could not be serialized
	ErrResultEncode = errors.New("integrations/operations: result encode failed")
	// ErrLinkFailed indicates a link operation failed
	ErrLinkFailed = errors.New("integrations/operations: link operation failed")
)
