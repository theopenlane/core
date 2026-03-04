package okta

import "errors"

var (
	// ErrCredentialsMissing indicates the org URL or API token is missing from the credential data
	ErrCredentialsMissing = errors.New("okta: org url or api token missing")
)
