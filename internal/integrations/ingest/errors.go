package ingest

import "errors"

// ErrMappingNotFound is returned when no mapping expression is available
var ErrMappingNotFound = errors.New("integration mapping not found")

// ErrMappingRequiredField is returned when a required mapping field is missing
var ErrMappingRequiredField = errors.New("mapping output missing required field")

// ErrMappingFilterType is returned when a filter expression does not return a boolean
var ErrMappingFilterType = errors.New("mapping filter did not return boolean")

// ErrMappingOutputEmpty is returned when a mapping expression returns nil
var ErrMappingOutputEmpty = errors.New("mapping output was empty")

// ErrDBClientRequired is returned when the database client is missing
var ErrDBClientRequired = errors.New("ingest: db client required")

// ErrIngestEmitterRequired is returned when the ingest event bus is missing
var ErrIngestEmitterRequired = errors.New("ingest: event emitter required")

// ErrIngestIntegrationRequired is returned when the integration id is missing
var ErrIngestIntegrationRequired = errors.New("ingest: integration id required")

// ErrIngestSchemaRequired is returned when the ingest schema is missing
var ErrIngestSchemaRequired = errors.New("ingest: schema required")

// ErrIngestSchemaUnsupported is returned when the ingest schema is unsupported
var ErrIngestSchemaUnsupported = errors.New("ingest: schema unsupported")

// ErrIngestProviderUnknown is returned when provider type is unknown
var ErrIngestProviderUnknown = errors.New("ingest: provider unknown")

// ErrMappingSchemaNotFound is returned when the required mapping schema is missing
var ErrMappingSchemaNotFound = errors.New("ingest: mapping schema not found")

// ErrExternalIDRequired is returned when an external ID is missing for persistence
var ErrExternalIDRequired = errors.New("ingest: external id required")

// ErrIngestOrgIDRequired is returned when the organization id is missing
var ErrIngestOrgIDRequired = errors.New("ingest: org id required")
