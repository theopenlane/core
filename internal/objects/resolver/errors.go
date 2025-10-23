package resolver

import "errors"

var (
	errUnsupportedProvider = errors.New("unsupported storage provider")
	errProviderDisabled    = errors.New("storage provider disabled")
)
