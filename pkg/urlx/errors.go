package urlx

import "errors"

var (
	// ErrTokenCreationFailed indicates that the token pair could not be created
	ErrTokenCreationFailed = errors.New("urlx: token creation failed")
	// ErrURLConstructionFailed indicates that the URL could not be constructed
	ErrURLConstructionFailed = errors.New("urlx: URL construction failed")
)
