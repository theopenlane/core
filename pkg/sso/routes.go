package sso

import (
	"fmt"
	"net/url"

	echo "github.com/theopenlane/echox"
)

// SSOLogin returns the path for the SSO login route with the organization ID query parameter
func SSOLogin(e *echo.Echo, orgID string) string {
	path := "/v1/sso/login"
	if e != nil {
		if p, err := e.Router().Routes().Reverse("SSOLogin"); err == nil {
			path = p
		}
	}

	if orgID == "" {
		return path
	}

	q := url.Values{}
	q.Set("organization_id", orgID)
	return fmt.Sprintf("%s?%s", path, q.Encode())
}

// SSOCallback returns the path for the SSO callback route
func SSOCallback(e *echo.Echo) string {
	if e != nil {
		if p, err := e.Router().Routes().Reverse("SSOCallback"); err == nil {
			return p
		}
	}
	return "/v1/sso/callback"
}
