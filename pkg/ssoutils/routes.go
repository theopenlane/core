package ssoutils

import (
	"fmt"
	"net/url"

	echo "github.com/theopenlane/echox"
)

// SSOLogin returns the path for the SSO login route with the organization ID query parameter
func SSOLogin(e *echo.Echo, orgID string) string {
	path := "/login/sso"

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

// SSOTokenAuthorize returns the path for the SSO token authorization route with token and org parameters
func SSOTokenAuthorize(e *echo.Echo, orgID, tokenID, tokenType string) string {
	path := "/v1/sso/token/authorize"

	if e != nil {
		if p, err := e.Router().Routes().Reverse("SSOTokenAuthorize"); err == nil {
			path = p
		}
	}

	q := url.Values{}

	if orgID != "" {
		q.Set("organization_id", orgID)
	}

	if tokenID != "" {
		q.Set("token_id", tokenID)
	}

	if tokenType != "" {
		q.Set("token_type", tokenType)
	}

	if len(q) == 0 {
		return path
	}

	return fmt.Sprintf("%s?%s", path, q.Encode())
}

// SSOTokenCallback returns the path for the SSO token callback route
func SSOTokenCallback(e *echo.Echo) string {
	if e != nil {
		if p, err := e.Router().Routes().Reverse("SSOTokenCallback"); err == nil {
			return p
		}
	}

	return "/v1/sso/token/callback"
}
