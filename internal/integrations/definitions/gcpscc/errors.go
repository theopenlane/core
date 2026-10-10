package gcpscc

import "errors"

var (
	// ErrProjectIDRequired indicates no project or organization ID was provided
	ErrProjectIDRequired = errors.New("gcpscc: project or organization ID required")
	// ErrProjectNumberRequired indicates no workload identity pool project number was provided
	ErrProjectNumberRequired = errors.New("gcpscc: workload identity project number required")
	// ErrServiceAccountKeyInvalid indicates the service account key JSON is invalid
	ErrServiceAccountKeyInvalid = errors.New("gcpscc: service account key invalid")
	// ErrSecurityCenterClientCreate indicates the SCC client could not be created
	ErrSecurityCenterClientCreate = errors.New("gcpscc: security center client creation failed")
	// ErrListSourcesFailed indicates the source listing request failed
	ErrListSourcesFailed = errors.New("gcpscc: list sources failed")
	// ErrListFindingsFailed indicates the findings listing request failed
	ErrListFindingsFailed = errors.New("gcpscc: list findings failed")
	// ErrFindingEncode indicates a finding payload could not be serialized
	ErrFindingEncode = errors.New("gcpscc: finding encode failed")
)
