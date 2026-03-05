package state

import "errors"

var (
	// ErrProviderStateDecode indicates provider state decoding failed
	ErrProviderStateDecode = errors.New("integration state provider decode failed")
	// ErrProviderStatePatchEncode indicates a provider state patch could not be encoded
	ErrProviderStatePatchEncode = errors.New("integration state patch encode failed")
)
