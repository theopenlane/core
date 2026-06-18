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

// withOverrides allows us to inject custom functions for testing
// isSSOEnforcedFunc and isSSOBypassedFunc, which are used to determine SSO enforcement
// and whether a user is exempt, respectively - seemed easier than
// spinning up a database + seeding information every time just to verify the
// behavior if a org setting is enforced or not, so creating some "mocking"
// functions for this purpose
func withOverrides(ssoFn func(context.Context, *generated.Client, string) (bool, error), bypassFn func(context.Context, *generated.Client, string, string, []string) bool) func() {
	origSSO := isSSOEnforcedFunc
	origBypass := isSSOBypassedFunc
	isSSOEnforcedFunc = ssoFn
	isSSOBypassedFunc = bypassFn
	return func() { isSSOEnforcedFunc = origSSO; isSSOBypassedFunc = origBypass }
}

func TestUnauthorizedRedirectToSSO(t *testing.T) {
	restore := withOverrides(
		func(context.Context, *generated.Client, string) (bool, error) { return true, nil },
		func(context.Context, *generated.Client, string, string, []string) bool { return false },
	)

	// the test temporarily overrides the isSSOEnforcedFunc and isSSOBypassedFunc
	// functions to simulate the behavior of SSO being enforced and the user
	// not being exempt, so we can test the redirect logic
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
		func(context.Context, *generated.Client, string) (bool, error) { return false, nil },
		func(context.Context, *generated.Client, string, string, []string) bool { return false },
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

func TestUnauthorizedOwnerBypass(t *testing.T) {
	restore := withOverrides(
		func(context.Context, *generated.Client, string) (bool, error) { return true, nil },
		func(context.Context, *generated.Client, string, string, []string) bool { return true },
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
