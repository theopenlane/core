package authentik

import "errors"

var (
	// ErrAPITokenMissing indicates the API token is missing from the credential
	ErrAPITokenMissing = errors.New("authentik: api token missing")
	// ErrBaseURLMissing indicates the Authentik base URL is missing from the credential
	ErrBaseURLMissing = errors.New("authentik: base url missing")
	// ErrHealthCheckFailed indicates the health check request failed
	ErrHealthCheckFailed = errors.New("authentik: health check failed")
	// ErrBrandFetchFailed indicates the default brand lookup request failed
	ErrBrandFetchFailed = errors.New("authentik: brand fetch failed")
	// ErrDefaultBrandMissing indicates the instance has no default brand to identify it by
	ErrDefaultBrandMissing = errors.New("authentik: default brand missing")
	// ErrDirectoryUsersFetchFailed indicates the users listing failed
	ErrDirectoryUsersFetchFailed = errors.New("authentik: directory users fetch failed")
	// ErrDirectoryGroupsFetchFailed indicates the groups listing failed
	ErrDirectoryGroupsFetchFailed = errors.New("authentik: directory groups fetch failed")
	// ErrPayloadEncode indicates a provider payload could not be serialized
	ErrPayloadEncode = errors.New("authentik: payload encode failed")
)
