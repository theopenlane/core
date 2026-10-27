package auth

import (
	"net/http"
	"slices"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
)

// BlockAnonymous is middleware that rejects anonymous JWT callers unless they hold one of the allowed capabilities
func BlockAnonymous(allowed ...auth.Capability) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			caller, ok := auth.CallerFromContext(c.Request().Context())
			if ok && caller.IsAnonymous() && !slices.ContainsFunc(allowed, caller.Has) {
				return echo.NewHTTPError(http.StatusUnauthorized, ErrAnonymousAccessNotAllowed.Error())
			}

			return next(c)
		}
	}
}
