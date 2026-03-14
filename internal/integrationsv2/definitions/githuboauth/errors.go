package githuboauth

import "errors"

var (
	// ErrOAuthTokenMissing indicates the OAuth access token is missing
	ErrOAuthTokenMissing = errors.New("githuboauth: oauth token missing")
	// ErrClientType indicates the provided client is not the expected type
	ErrClientType = errors.New("githuboauth: unexpected client type")
	// ErrAPIRequest indicates a GitHub API request error
	ErrAPIRequest = errors.New("githuboauth: api request error")
	// ErrRepositoryInvalid indicates a repository name is not in owner/repo format
	ErrRepositoryInvalid = errors.New("githuboauth: invalid repository name (expected owner/repo)")
)
