package fossa

import "errors"

var (
	// ErrAPITokenMissing indicates the installation credential has no FOSSA API token
	ErrAPITokenMissing = errors.New("fossa: api token is required")
	// ErrClientBuild indicates the FOSSA API client could not be constructed
	ErrClientBuild = errors.New("fossa: unable to build api client")
	// ErrUnauthorized indicates FOSSA rejected the API token
	ErrUnauthorized = errors.New("fossa: api token was rejected, a full api token is required (push-only tokens cannot read issues)")
	// ErrRateLimited indicates FOSSA rate limited the request
	ErrRateLimited = errors.New("fossa: rate limited by the api")
	// ErrAPIRequest indicates the FOSSA API returned an unexpected status
	ErrAPIRequest = errors.New("fossa: unexpected api response")
	// ErrIssuesFetchFailed indicates issues could not be listed
	ErrIssuesFetchFailed = errors.New("fossa: unable to list issues")
	// ErrIssueEncode indicates an issue payload could not be encoded into an ingest envelope
	ErrIssueEncode = errors.New("fossa: unable to encode issue payload")
	// ErrCategoriesFetchFailed indicates the issue category counts could not be retrieved
	ErrCategoriesFetchFailed = errors.New("fossa: unable to fetch issue categories")
	// ErrOrganizationFetchFailed indicates the organization details could not be retrieved
	ErrOrganizationFetchFailed = errors.New("fossa: unable to fetch organization details")
	// ErrOrganizationIDMissing indicates the organization response carried no ID for the installation identity
	ErrOrganizationIDMissing = errors.New("fossa: organization id missing")
)
