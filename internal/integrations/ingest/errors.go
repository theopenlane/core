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

// ErrMappingSchemaNotFound is returned when the required mapping schema is missing
var ErrMappingSchemaNotFound = errors.New("ingest: mapping schema not found")

// ErrExternalIDRequired is returned when an external ID is missing for persistence
var ErrExternalIDRequired = errors.New("ingest: external id required")
