package okta

import "errors"

var (
	// ErrAPITokenMissing indicates the Okta API token is missing from the credential
	ErrAPITokenMissing = errors.New("okta: api token missing")
	// ErrOrgURLMissing indicates the Okta org URL is missing from the credential
	ErrOrgURLMissing = errors.New("okta: org url missing")
	// ErrClientType indicates the provided client is not an Okta API client
	ErrClientType = errors.New("okta: unexpected client type")
)
