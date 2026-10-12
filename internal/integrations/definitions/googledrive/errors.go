package googledrive

import "errors"

var (
	// ErrOAuthTokenMissing indicates the OAuth access token is missing
	ErrOAuthTokenMissing = errors.New("googledrive: oauth token missing")
	// ErrDriveServiceBuildFailed indicates the Drive SDK client could not be constructed
	ErrDriveServiceBuildFailed = errors.New("googledrive: drive service build failed")
	// ErrHealthCheckFailed indicates the health check request failed
	ErrHealthCheckFailed = errors.New("googledrive: health check failed")
	// ErrCredentialEncode indicates the credential could not be serialized
	ErrCredentialEncode = errors.New("googledrive: credential encode failed")
	// ErrAccountUnresolved indicates the Drive About API did not report the connected account
	ErrAccountUnresolved = errors.New("googledrive: account unresolved")
	// ErrFolderIDMissing indicates the folder ID is not configured
	ErrFolderIDMissing = errors.New("googledrive: folder id missing from user input")
	// ErrFolderListFailed indicates the Drive folder listing request failed
	ErrFolderListFailed = errors.New("googledrive: folder list failed")
	// ErrPayloadEncode indicates a provider payload could not be serialized
	ErrPayloadEncode = errors.New("googledrive: payload encode failed")
)
