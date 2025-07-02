package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	echo "github.com/theopenlane/echox"
	iamauth "github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/utils/ulids"

	generated "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	dbtest "github.com/theopenlane/core/pkg/testutils/fga"
)

func TestAPITokenSSOEnforcement(t *testing.T) {
	client := dbtest.NewPostgresClient(t)

	ctx := privacy.DecisionContext(context.Background(), privacy.Allow)
	ctx = generated.NewContext(ctx, client)

	ownerID := ulids.New().String()
	ownerSetting, err := client.UserSetting.Create().Save(ctx)
	require.NoError(t, err)
	_, err = client.User.Create().
		SetID(ownerID).
		SetEmail("owner@example.com").
		SetDisplayName("owner").
		SetPassword("p@$$w0rd$@r3gr3@t!").
		SetSettingID(ownerSetting.ID).
		Save(ctx)
	require.NoError(t, err)

	orgSetting, err := client.OrganizationSetting.Create().
		SetBillingEmail("org@example.com").
		Save(ctx)
	require.NoError(t, err)

	org, err := client.Organization.Create().
		SetName("org").
		SetSettingID(orgSetting.ID).
		Save(ctx)
	require.NoError(t, err)

	_, err = orgSetting.Update().
		SetOrganizationID(org.ID).
		Save(ctx)
	require.NoError(t, err)

	token := client.APIToken.Create().
		SetOwnerID(org.ID).
		SetName("token").
		SaveX(ctx)

	validator := &tokens.MockValidator{
		OnParse: func(string) (*tokens.Claims, error) {
			return &tokens.Claims{OrgID: org.ID, UserID: ownerID}, nil
		},
		OnVerify: func(string) (*tokens.Claims, error) { return nil, nil },
	}

	conf := NewAuthOptions(
		WithDBClient(client),
		WithValidator(validator),
	)

	e := echo.New()
	e.Use(Authenticate(&conf))
	e.GET("/", func(c echo.Context) error { return c.NoContent(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(iamauth.Authorization, "Bearer "+token.Token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	_, err = orgSetting.Update().
		SetIdentityProviderLoginEnforced(true).
		Save(ctx)
	require.NoError(t, err)

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set(iamauth.Authorization, "Bearer "+token.Token)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusUnauthorized, rec2.Code)

	client.APIToken.UpdateOneID(token.ID).
		SetSSOAuthorizations(models.SSOAuthorizationMap{org.ID: time.Now()}).
		ExecX(ctx)

	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	req3.Header.Set(iamauth.Authorization, "Bearer "+token.Token)
	rec3 := httptest.NewRecorder()
	e.ServeHTTP(rec3, req3)
	assert.Equal(t, http.StatusOK, rec3.Code)
}

func TestPATTokenSSOEnforcement(t *testing.T) {
	client := dbtest.NewPostgresClient(t)

	ctx := privacy.DecisionContext(context.Background(), privacy.Allow)
	ctx = generated.NewContext(ctx, client)

	userID := ulids.New().String()
	setting, err := client.UserSetting.Create().Save(ctx)
	require.NoError(t, err)
	user, err := client.User.Create().
		SetID(userID).
		SetEmail("user@example.com").
		SetDisplayName("user").
		SetPassword("p@$$w0rd$@r3gr3@t!").
		SetSettingID(setting.ID).
		Save(ctx)
	require.NoError(t, err)

	orgSetting, err := client.OrganizationSetting.Create().
		SetBillingEmail("org@example.com").
		Save(ctx)
	require.NoError(t, err)

	org, err := client.Organization.Create().
		SetName("org").
		SetSettingID(orgSetting.ID).
		Save(ctx)
	require.NoError(t, err)

	_, err = orgSetting.Update().
		SetOrganizationID(org.ID).
		Save(ctx)
	require.NoError(t, err)

	role := enums.RoleMember
	client.OrgMembership.Create().SetInput(generated.CreateOrgMembershipInput{
		OrganizationID: org.ID,
		UserID:         user.ID,
		Role:           &role,
	}).ExecX(ctx)

	pat := client.PersonalAccessToken.Create().
		SetOwnerID(user.ID).
		AddOrganizationIDs(org.ID).
		SetName("p").
		SaveX(ctx)

	validator := &tokens.MockValidator{
		OnParse: func(string) (*tokens.Claims, error) {
			return &tokens.Claims{OrgID: org.ID, UserID: user.ID}, nil
		},
		OnVerify: func(string) (*tokens.Claims, error) { return nil, nil },
	}

	conf := NewAuthOptions(
		WithDBClient(client),
		WithValidator(validator),
	)

	e := echo.New()
	e.Use(Authenticate(&conf))
	e.GET("/", func(c echo.Context) error { return c.NoContent(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(iamauth.Authorization, "Bearer "+pat.Token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	_, err = orgSetting.Update().
		SetIdentityProviderLoginEnforced(true).
		Save(ctx)
	require.NoError(t, err)

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set(iamauth.Authorization, "Bearer "+pat.Token)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusUnauthorized, rec2.Code)

	client.PersonalAccessToken.UpdateOneID(pat.ID).
		SetSSOAuthorizations(models.SSOAuthorizationMap{org.ID: time.Now()}).
		ExecX(ctx)

	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	req3.Header.Set(iamauth.Authorization, "Bearer "+pat.Token)
	rec3 := httptest.NewRecorder()
	e.ServeHTTP(rec3, req3)
	assert.Equal(t, http.StatusOK, rec3.Code)
}
