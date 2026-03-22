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
	// ErrSecurityCenterClientCreate indicates the SCC client could not be created
	ErrSecurityCenterClientCreate = errors.New("gcpscc: security center client creation failed")
	// ErrOperationConfigInvalid indicates operation config could not be decoded
	ErrOperationConfigInvalid = errors.New("gcpscc: operation config invalid")
	// ErrListSourcesFailed indicates the source listing request failed
	ErrListSourcesFailed = errors.New("gcpscc: list sources failed")
	// ErrListFindingsFailed indicates the findings listing request failed
	ErrListFindingsFailed = errors.New("gcpscc: list findings failed")
	// ErrNotificationConfigScanFailed indicates the notification config scan failed
	ErrNotificationConfigScanFailed = errors.New("gcpscc: notification config scan failed")
	// ErrFindingEncode indicates a finding payload could not be serialized
	ErrFindingEncode = errors.New("gcpscc: finding encode failed")
	// ErrResultEncode indicates an operation result could not be serialized
	ErrResultEncode = errors.New("gcpscc: result encode failed")
)
