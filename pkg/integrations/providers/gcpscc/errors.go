package gcpscc

import "errors"

var (
	// ErrSecurityCenterClientRequired indicates the security center client was not provided or is invalid
	ErrSecurityCenterClientRequired = errors.New("gcpscc: security center client required")
	// ErrBeginAuthNotSupported indicates BeginAuth is not supported for GCP SCC providers
	ErrBeginAuthNotSupported = errors.New("gcpscc: BeginAuth is not supported; supply metadata via credential schema")
)
