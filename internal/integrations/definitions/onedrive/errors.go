package onedrive

import "errors"

var (
	// ErrOAuthTokenMissing indicates the OAuth access token is missing
	ErrOAuthTokenMissing = errors.New("onedrive: oauth token missing")
	// ErrClientBuildFailed indicates the Graph SDK client could not be constructed
	ErrClientBuildFailed = errors.New("onedrive: client build failed")
	// ErrHealthCheckFailed indicates the health check request failed
	ErrHealthCheckFailed = errors.New("onedrive: health check failed")
	// ErrCredentialEncode indicates the credential could not be serialized
	ErrCredentialEncode = errors.New("onedrive: credential encode failed")
	// ErrCredentialDecode indicates the credential could not be deserialized
	ErrCredentialDecode = errors.New("onedrive: credential decode failed")
	// ErrResultEncode indicates an operation result could not be serialized
	ErrResultEncode = errors.New("onedrive: result encode failed")
	// ErrExportFailed indicates the file download request failed
	ErrExportFailed = errors.New("onedrive: file export failed")
	// ErrFolderIDMissing indicates the folder ID is not configured
	ErrFolderIDMissing = errors.New("onedrive: folder id missing from user input")
	// ErrFolderListFailed indicates the folder children listing request failed
	ErrFolderListFailed = errors.New("onedrive: folder list failed")
	// ErrPayloadEncode indicates a provider payload could not be serialized
	ErrPayloadEncode = errors.New("onedrive: payload encode failed")
)
