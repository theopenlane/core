package github

import "errors"

var (
	// ErrAPIRequest indicates a GitHub API request failed with a non-2xx status
	ErrAPIRequest = errors.New("github: api request failed")
	// ErrOAuthTokenMissing indicates the OAuth token is not present in the credential payload
	ErrOAuthTokenMissing = errors.New("github: oauth token missing")
	// ErrAccessTokenEmpty indicates the access token field is empty
	ErrAccessTokenEmpty = errors.New("github: access token empty")
)
