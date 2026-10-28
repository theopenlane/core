package auth

import (
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
)

// BlockNonTrustCenterAnonymous is middleware that rejects anonymous JWT callers
// that are not trust center tokens. Trust center anon callers (CapTrustCenterAnonymous)
// and all regular authenticated users are allowed through
// This prevents questionnaires and other non-user JWTs from
// access the API data with the JWT
func BlockNonTrustCenterAnonymous() echo.MiddlewareFunc {
	return BlockAnonymous(auth.CapTrustCenterAnonymous)
}
