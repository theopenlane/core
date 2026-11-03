package auth

import (
	"net/http"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
)

// BlockTFAEnrollmentRequired is middleware that rejects callers holding an enrollment-only session
func BlockTFAEnrollmentRequired() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if auth.IsTFAEnrollmentRequiredFromContext(c.Request().Context()) {
				return echo.NewHTTPError(http.StatusUnauthorized, auth.ErrTFAEnrollmentRequired.Error())
			}

			return next(c)
		}
	}
}
