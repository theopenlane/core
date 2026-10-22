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
	// ErrNoDocumentsSelected indicates neither documents nor a folder are configured for the sync
	ErrNoDocumentsSelected = errors.New("googledrive: no documents selected")
	// ErrFolderListFailed indicates the Drive folder listing request failed
	ErrFolderListFailed = errors.New("googledrive: folder list failed")
	// ErrInstallationMetadataDecode indicates the installation metadata could not be decoded
	ErrInstallationMetadataDecode = errors.New("googledrive: installation metadata decode failed")
	// ErrPayloadEncode indicates a provider payload could not be serialized
	ErrPayloadEncode = errors.New("googledrive: payload encode failed")
	// ErrPickerLegacyInstallation indicates the Picker was requested for an installation still on the legacy drive.readonly app
	ErrPickerLegacyInstallation = errors.New("googledrive: picker is unavailable for legacy installations, reconnect the integration to select documents")
	// ErrPickerTokenFailed indicates an access token for the Picker could not be obtained
	ErrPickerTokenFailed = errors.New("googledrive: picker token unavailable")
	// ErrPickerAppIDUnresolved indicates the Google Cloud project number could not be read from the configured OAuth client ID
	ErrPickerAppIDUnresolved = errors.New("googledrive: picker app id could not be derived from the client id")
)
