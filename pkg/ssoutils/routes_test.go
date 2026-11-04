package ssoutils

import (
	"net/http"
	"testing"

	echo "github.com/theopenlane/echox"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
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
	assert.Check(t, is.Equal(want, SSOLogin(e, orgID)))
	assert.Check(t, is.Equal("/v1/sso/login", SSOLogin(e, "")))

	e2 := echo.New()
	assert.Check(t, is.Equal(want, SSOLogin(e2, orgID)))

	assert.Check(t, is.Equal(want, SSOLogin(nil, orgID)))
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

	assert.Check(t, is.Equal("/v1/sso/callback", SSOCallback(e)))

	e2 := echo.New()
	assert.Check(t, is.Equal("/v1/sso/callback", SSOCallback(e2)))

	assert.Check(t, is.Equal("/v1/sso/callback", SSOCallback(nil)))
}

func TestSSOLoginGroupRegisteredRoute(t *testing.T) {
	t.Parallel()

	e := echo.New()
	g := e.Group("v1")

	_, err := g.AddRoute(echo.Route{
		Name:   "SSOLogin",
		Method: http.MethodPost,
		Path:   "/sso/login",
		Handler: func(c echo.Context) error {
			return nil
		},
	})
	assert.Check(t, err)

	const orgID = "abc123"
	assert.Check(t, is.Equal("/v1/sso/login?organization_id="+orgID, SSOLogin(e, orgID)))
}
