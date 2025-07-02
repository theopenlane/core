package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/echox"
	iamauth "github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	dbtest "github.com/theopenlane/core/pkg/testutils/fga"
)

func TestUnauthorizedRedirectToSSO(t *testing.T) {
	client := dbtest.NewPostgresClient(t)

	ctx := privacy.DecisionContext(context.Background(), privacy.Allow)

	setting, err := client.OrganizationSetting.Create().
		SetBillingEmail("org1@example.com").
		Save(ctx)
	require.NoError(t, err)

	_, err = client.Organization.Create().
		SetName("org1").
		SetSettingID(setting.ID).
		Save(ctx)
	require.NoError(t, err)

	_, err = setting.Update().
		SetIdentityProviderLoginEnforced(true).
		Save(ctx)
	require.NoError(t, err)

	validator := &tokens.MockValidator{
		OnParse: func(string) (*tokens.Claims, error) {
			return &tokens.Claims{OrgID: "org1"}, nil
		},
	}

	conf := NewAuthOptions(
		WithDBClient(client),
		WithValidator(validator),
	)

	e := echox.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(iamauth.Authorization, "Bearer token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = unauthorized(c, errors.New("invalid"), &conf, validator)
	require.NoError(t, err)
	require.Equal(t, http.StatusFound, rec.Code)
	require.Equal(t, "/v1/sso/login?organization_id=org1", rec.Header().Get("Location"))
}

func TestUnauthorizedNoSSORedirect(t *testing.T) {
	client := dbtest.NewPostgresClient(t)

	ctx := privacy.DecisionContext(context.Background(), privacy.Allow)

	setting, err := client.OrganizationSetting.Create().
		SetBillingEmail("org2@example.com").
		Save(ctx)
	require.NoError(t, err)

	_, err = client.Organization.Create().
		SetName("org2").
		SetSettingID(setting.ID).
		Save(ctx)
	require.NoError(t, err)

	_, err = setting.Update().
		SetIdentityProviderLoginEnforced(false).
		Save(ctx)
	require.NoError(t, err)

	validator := &tokens.MockValidator{
		OnParse: func(string) (*tokens.Claims, error) {
			return &tokens.Claims{OrgID: "org2"}, nil
		},
	}

	conf := NewAuthOptions(
		WithDBClient(client),
		WithValidator(validator),
	)

	e := echox.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(iamauth.Authorization, "Bearer token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = unauthorized(c, errors.New("invalid"), &conf, validator)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}
