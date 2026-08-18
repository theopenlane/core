package urlx

import "errors"

var (
	// ErrTokenCreationFailed indicates that the token pair could not be created
	ErrTokenCreationFailed = errors.New("urlx: token creation failed")
	// ErrURLConstructionFailed indicates that the URL could not be constructed
	ErrURLConstructionFailed = errors.New("urlx: URL construction failed")
	// ErrEmptyURL indicates the input URL is empty or whitespace
	ErrEmptyURL = errors.New("urlx: url is required")
	// ErrInvalidURL indicates the input could not be parsed as a URL
	ErrInvalidURL = errors.New("urlx: invalid url")
	// ErrMissingHost indicates the parsed URL does not contain a host
	ErrMissingHost = errors.New("urlx: url host is required")
	// ErrUnsupportedScheme indicates the URL scheme is not http or https
	ErrUnsupportedScheme = errors.New("urlx: unsupported url scheme")
	// ErrSizeLimitExceeded indicates a response body exceeded the configured size limit
	ErrSizeLimitExceeded = errors.New("urlx: response size limit exceeded")
)
