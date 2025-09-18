package objects

import "errors"

var (
	// ErrNoStorageProvider is returned when no storage provider is available
	ErrNoStorageProvider = errors.New("no storage provider available")
	// ErrInsufficientProviderInfo is returned when insufficient information to resolve storage client
	ErrInsufficientProviderInfo = errors.New("insufficient information to resolve storage client: need integration_id+hush_id or organization_id")
	// ErrProviderHintsRequired is returned when provider hints are required for file upload
	ErrProviderHintsRequired = errors.New("provider hints required for file upload")
)
