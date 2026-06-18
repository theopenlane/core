package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/echox"
	iamauth "github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"

	generated "github.com/theopenlane/core/internal/ent/generated"
	sso "github.com/theopenlane/core/pkg/ssoutils"
)

// withOverrides injects a custom userMustSSOFunc for testing. userMustSSOFunc encapsulates the
// full SSO routing decision (enforcement, owner, per-user and per-domain exemptions), so overriding
// it lets us verify the unauthorized redirect behavior without a database
func withOverrides(mustSSOFn func(context.Context, *generated.Client, string, string) (bool, error)) func() {
	origMustSSO := userMustSSOFunc
	userMustSSOFunc = mustSSOFn
	return func() { userMustSSOFunc = origMustSSO }
}

func TestUnauthorizedRedirectToSSO(t *testing.T) {
	restore := withOverrides(
		func(context.Context, *generated.Client, string, string) (bool, error) { return true, nil },
	)

	// the test temporarily overrides the isSSOEnforcedFunc and orgRoleFunc
	// functions to simulate the behavior of SSO being enforced and the user
	// being a member of the organization, so we can test the redirect logic
	// without needing a real database or organization setup
	// this restore function will reset the overrides after the test is done
	defer restore()

	validator := &tokens.MockValidator{
		OnParse: func(string) (*tokens.Claims, error) { return &tokens.Claims{OrgID: "org", UserID: "user"}, nil },
	}

	conf := NewAuthOptions(
		WithDBClient(&generated.Client{}),
		WithValidator(validator),
	)

	e := echox.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(iamauth.Authorization, "Bearer token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// we basically want to simulate that the unauthorized function is called
	// and what the behavior is when it is called
	err := unauthorized(c, errors.New("invalid"), &conf, validator)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code)
	// the "location" is the destination we'd be 302'ing someone to, so it should match the
	// helper function SSOLogin
	assert.Equal(t, sso.SSOLogin(e, "org"), rec.Header().Get("Location"))
}

func TestUnauthorizedNoSSORedirect(t *testing.T) {
	restore := withOverrides(
		func(context.Context, *generated.Client, string, string) (bool, error) { return false, nil },
	)

	defer restore()

	validator := &tokens.MockValidator{
		OnParse: func(string) (*tokens.Claims, error) { return &tokens.Claims{OrgID: "org", UserID: "user"}, nil },
	}

	conf := NewAuthOptions(
		WithDBClient(&generated.Client{}),
		WithValidator(validator),
	)

	e := echox.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(iamauth.Authorization, "Bearer token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := unauthorized(c, errors.New("invalid"), &conf, validator)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TestUnauthorizedExemptNoRedirect verifies an exempt user (e.g. owner or per-user exemption),
// for whom userMustSSO resolves false, is not redirected through SSO
func TestUnauthorizedExemptNoRedirect(t *testing.T) {
	restore := withOverrides(
		func(context.Context, *generated.Client, string, string) (bool, error) { return false, nil },
	)

	defer restore()

	validator := &tokens.MockValidator{
		OnParse: func(string) (*tokens.Claims, error) { return &tokens.Claims{OrgID: "org", UserID: "owner"}, nil },
	}

	conf := NewAuthOptions(
		WithDBClient(&generated.Client{}),
		WithValidator(validator),
	)

	e := echox.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(iamauth.Authorization, "Bearer token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := unauthorized(c, errors.New("invalid"), &conf, validator)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
