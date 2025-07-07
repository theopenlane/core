package ssoutils

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	echo "github.com/theopenlane/echox"
)

func TestSSOLogin(t *testing.T) {
	t.Parallel()

	e := echo.New()
	e.AddRoute(echo.Route{
		Name:   "SSOLogin",
		Method: http.MethodGet,
		Path:   "/v1/sso/login",
		Handler: func(c echo.Context) error {
			return nil
		},
	})

	const orgID = "abc123"
	want := "/v1/sso/login?organization_id=" + orgID
	assert.Equal(t, want, SSOLogin(e, orgID))
	assert.Equal(t, "/v1/sso/login", SSOLogin(e, ""))

	e2 := echo.New()
	assert.Equal(t, want, SSOLogin(e2, orgID))

	assert.Equal(t, want, SSOLogin(nil, orgID))
}

func TestSSOCallback(t *testing.T) {
	t.Parallel()

	e := echo.New()
	e.AddRoute(echo.Route{
		Name:   "SSOCallback",
		Method: http.MethodGet,
		Path:   "/v1/sso/callback",
		Handler: func(c echo.Context) error {
			return nil
		},
	})

	assert.Equal(t, "/v1/sso/callback", SSOCallback(e))

	e2 := echo.New()
	assert.Equal(t, "/v1/sso/callback", SSOCallback(e2))

	assert.Equal(t, "/v1/sso/callback", SSOCallback(nil))
}
