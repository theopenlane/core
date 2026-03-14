package googleworkspace

import "errors"

var (
	// ErrOAuthTokenMissing indicates the OAuth access token is missing
	ErrOAuthTokenMissing = errors.New("googleworkspace: oauth token missing")
	// ErrClientType indicates the provided client is not the expected type
	ErrClientType = errors.New("googleworkspace: unexpected client type")
)
