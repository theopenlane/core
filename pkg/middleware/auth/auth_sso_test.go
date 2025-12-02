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

	"github.com/theopenlane/core/pkg/enums"
	sso "github.com/theopenlane/core/pkg/ssoutils"
	generated "github.com/theopenlane/ent/generated"
)

// withOverrides allows us to inject custom functions for testing
// isSSOEnforcedFunc and orgRoleFunc, which are used to determine SSO enforcement
// and the user's role in the organization, respectively - seemed easier than
// spinning up a database + seeding information every time just to verify the
// behavior if a org setting is enforced or not, so creating some "mocking"
// functions for this purpose
func withOverrides(ssoFn func(context.Context, *generated.Client, string) (bool, error), roleFn func(context.Context, *generated.Client, string, string) (enums.Role, error)) func() {
	origSSO := isSSOEnforcedFunc
	origRole := orgRoleFunc
	isSSOEnforcedFunc = ssoFn
	orgRoleFunc = roleFn
	return func() { isSSOEnforcedFunc = origSSO; orgRoleFunc = origRole }
}

func TestUnauthorizedRedirectToSSO(t *testing.T) {
	restore := withOverrides(
		func(context.Context, *generated.Client, string) (bool, error) { return true, nil },
		func(context.Context, *generated.Client, string, string) (enums.Role, error) {
			return enums.RoleMember, nil
		},
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
		func(context.Context, *generated.Client, string) (bool, error) { return false, nil },
		func(context.Context, *generated.Client, string, string) (enums.Role, error) {
			return enums.RoleMember, nil
		},
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
		func(context.Context, *generated.Client, string, string) (enums.Role, error) {
			return enums.RoleOwner, nil
		},
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
