package okta

import "errors"

var (
	// ErrCredentialsMissing indicates the org URL or API token is missing from the credential data
	ErrCredentialsMissing = errors.New("okta: org url or api token missing")
	// ErrAPIRequest indicates an Okta API request failed with a non-2xx status
	ErrAPIRequest = errors.New("okta: api request failed")
)
