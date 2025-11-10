package keymaker

import "errors"

var (
	ErrMissingOrgID         = errors.New("keymaker: missing org id")
	ErrMissingIntegration   = errors.New("keymaker: missing integration id")
	ErrMissingProvider      = errors.New("keymaker: missing provider id")
	ErrFactoryNotRegistered = errors.New("keymaker: client factory not registered")
	ErrRefreshUnsupported   = errors.New("keymaker: refresh unsupported")
	ErrRefreshTokenMissing  = errors.New("keymaker: refresh token missing")
)
