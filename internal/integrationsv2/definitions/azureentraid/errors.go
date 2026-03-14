package azureentraid

import "errors"

var (
	// ErrOAuthTokenMissing indicates the OAuth access token is missing
	ErrOAuthTokenMissing = errors.New("azureentraid: oauth token missing")
	// ErrClientType indicates the provided client is not the expected type
	ErrClientType = errors.New("azureentraid: unexpected client type")
)
