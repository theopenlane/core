package proxy

import "errors"

var (
	// ErrMissingFileID indicates that the file ID is missing
	ErrMissingFileID = errors.New("missing file id")
	// ErrMissingObjectURI indicates that the file metadata is missing the object URI
	ErrMissingObjectURI = errors.New("file metadata missing object URI")
	// ErrInvalidSecretLength indicates that the secret must be 128 bytes when provided
	ErrInvalidSecretLength = errors.New("secret must be 128 bytes when provided")
	// ErrCallerRequired indicates that a caller is required in the context.
	ErrCallerRequired = errors.New("caller required in context")
	// ErrShareVerifiedTokenInvalid indicates the share-verified token failed signature, issuer, audience, or expiry validation
	ErrShareVerifiedTokenInvalid = errors.New("invalid share verified token")
	// ErrShareVerifiedTokenType indicates the token is not a share-verified token
	ErrShareVerifiedTokenType = errors.New("share verified token type mismatch")
)
