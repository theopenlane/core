package types

import "errors"

var (
	// ErrIngestOrgIDRequired is returned when the organization id is missing from an ingest request
	ErrIngestOrgIDRequired = errors.New("ingest: org id required")
	// ErrIngestIntegrationRequired is returned when the integration id is missing from an ingest request
	ErrIngestIntegrationRequired = errors.New("ingest: integration id required")
	// ErrIngestProviderUnknown is returned when the provider type is unknown in an ingest request
	ErrIngestProviderUnknown = errors.New("ingest: provider unknown")
	// ErrIngestOperationRequired is returned when the operation name is missing from an ingest request
	ErrIngestOperationRequired = errors.New("ingest: operation required")
	// ErrProviderStateDecode is returned when provider state decoding fails
	ErrProviderStateDecode = errors.New("integration state provider decode failed")
)
