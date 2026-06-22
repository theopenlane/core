package auth

import (
	"net/http"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
)

// BlockNonTrustCenterAnonymous is middleware that rejects anonymous JWT callers
// that are not trust center tokens. Trust center anon callers (CapTrustCenterAnonymous)
// and all regular authenticated users are allowed through
// This prevents questionnaires and other non-user JWTs from
// access the API data with the JWT
func BlockNonTrustCenterAnonymous() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			caller, ok := auth.CallerFromContext(c.Request().Context())
			if ok && caller != nil && caller.OrganizationRole == auth.AnonymousRole {
				if !caller.Has(auth.CapTrustCenterAnonymous) {
					return echo.NewHTTPError(http.StatusUnauthorized, ErrAnonymousAccessNotAllowed.Error())
				}
			}

			return next(c)
		}
	}
}
