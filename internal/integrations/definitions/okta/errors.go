package okta

import "errors"

var (
	// ErrAPITokenMissing indicates the Okta API token is missing from the credential
	ErrAPITokenMissing = errors.New("okta: api token missing")
	// ErrOrgURLMissing indicates the Okta org URL is missing from the credential
	ErrOrgURLMissing = errors.New("okta: org url missing")
	// ErrClientConfigInvalid indicates the Okta client configuration is invalid
	ErrClientConfigInvalid = errors.New("okta: client config invalid")
	// ErrUserLookupFailed indicates the current user lookup failed
	ErrUserLookupFailed = errors.New("okta: user lookup failed")
	// ErrOrgSettingsFetchFailed indicates the org settings lookup for the tenant identity failed
	ErrOrgSettingsFetchFailed = errors.New("okta: org settings fetch failed")
	// ErrOrgIDMissing indicates the org settings response carried no org id for the tenant identity
	ErrOrgIDMissing = errors.New("okta: org id missing")
	// ErrDirectoryUsersFetchFailed indicates the Okta users listing failed
	ErrDirectoryUsersFetchFailed = errors.New("okta: directory users fetch failed")
	// ErrDirectoryGroupsFetchFailed indicates the Okta groups listing failed
	ErrDirectoryGroupsFetchFailed = errors.New("okta: directory groups fetch failed")
	// ErrDirectoryGroupMembersFetchFailed indicates group member listing failed
	ErrDirectoryGroupMembersFetchFailed = errors.New("okta: directory group members fetch failed")
	// ErrPayloadEncode indicates a directory payload could not be serialized
	ErrPayloadEncode = errors.New("okta: payload encode failed")
)
