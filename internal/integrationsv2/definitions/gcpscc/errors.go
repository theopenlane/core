package gcpscc

import "errors"

var (
	// ErrClientType indicates the provided client is not the expected type
	ErrClientType = errors.New("gcpscc: unexpected client type")
	// ErrCredentialMetadataRequired indicates no credential metadata was provided
	ErrCredentialMetadataRequired = errors.New("gcpscc: credential metadata required")
	// ErrMetadataDecode indicates credential metadata could not be decoded
	ErrMetadataDecode = errors.New("gcpscc: failed to decode credential metadata")
	// ErrProjectIDRequired indicates no project or organization ID was provided
	ErrProjectIDRequired = errors.New("gcpscc: project or organization ID required")
	// ErrSourceIDRequired indicates no SCC source ID was provided
	ErrSourceIDRequired = errors.New("gcpscc: source ID required")
	// ErrServiceAccountKeyInvalid indicates the service account key JSON is invalid
	ErrServiceAccountKeyInvalid = errors.New("gcpscc: service account key invalid")
	// ErrAccessTokenMissing indicates no access token or service account key is available
	ErrAccessTokenMissing = errors.New("gcpscc: no access token or service account key")
	// ErrSecurityCenterClientCreate indicates the SCC client could not be created
	ErrSecurityCenterClientCreate = errors.New("gcpscc: security center client creation failed")
)
