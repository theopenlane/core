package cloudflare

import "errors"

var (
	// ErrTokenVerificationFailed indicates the Cloudflare API returned errors during token verification
	ErrTokenVerificationFailed = errors.New("cloudflare: token verification returned errors")
)
