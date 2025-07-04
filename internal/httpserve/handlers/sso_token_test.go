package handlers_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/sso"
)

func (suite *HandlerTestSuite) TestSSOTokenAuthorizeHandler() {
	t := suite.T()

	suite.e.GET("/v1/sso/token/authorize", suite.h.SSOTokenAuthorizeHandler)

	ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)
	ctx = ent.NewContext(ctx, suite.db)

	setting := suite.db.OrganizationSetting.Create().SetInput(ent.CreateOrganizationSettingInput{
		IdentityProviderLoginEnforced: lo.ToPtr(true),
		IdentityProvider:              lo.ToPtr(enums.SSOProviderOkta),
		IdentityProviderClientID:      lo.ToPtr("id"),
		IdentityProviderClientSecret:  lo.ToPtr("secret"),
		OidcDiscoveryEndpoint:         lo.ToPtr("http://example.com"),
	}).SaveX(ctx)

	org := suite.db.Organization.Create().SetInput(ent.CreateOrganizationInput{
		Name:      gofakeit.Name(),
		SettingID: &setting.ID,
	}).SaveX(ctx)

	token := suite.db.APIToken.Create().SetOwnerID(org.ID).SetName("t").SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, sso.SSOTokenAuthorize(suite.e, org.ID, token.ID, "api"), nil)
	rec := httptest.NewRecorder()

	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusFound, rec.Code)
}

func (suite *HandlerTestSuite) TestSSOTokenCallbackHandler() {
	t := suite.T()

	suite.e.GET("/v1/sso/token/callback", suite.h.SSOTokenCallbackHandler)

	ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)
	ctx = ent.NewContext(ctx, suite.db)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"tok","token_type":"Bearer"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	setting := suite.db.OrganizationSetting.Create().SetInput(ent.CreateOrganizationSettingInput{
		IdentityProviderLoginEnforced: lo.ToPtr(true),
		IdentityProvider:              lo.ToPtr(enums.SSOProviderOkta),
		IdentityProviderClientID:      lo.ToPtr("id"),
		IdentityProviderClientSecret:  lo.ToPtr("secret"),
		OidcDiscoveryEndpoint:         lo.ToPtr(ts.URL),
	}).SaveX(ctx)

	org := suite.db.Organization.Create().SetInput(ent.CreateOrganizationInput{
		Name:      gofakeit.Name(),
		SettingID: &setting.ID,
	}).SaveX(ctx)

	token := suite.db.APIToken.Create().SetOwnerID(org.ID).SetName("t").SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, sso.SSOTokenCallback(suite.e)+"?state=s&code=c", nil)
	for _, c := range []*http.Cookie{
		{Name: "token_id", Value: token.ID, Path: "/"},
		{Name: "token_type", Value: "api", Path: "/"},
		{Name: "organization_id", Value: org.ID, Path: "/"},
		{Name: "state", Value: "s", Path: "/"},
		{Name: "nonce", Value: "n", Path: "/"},
	} {
		req.AddCookie(c)
	}

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	refreshed, err := suite.db.APIToken.Get(ctx, token.ID)
	require.NoError(t, err)
	require.NotNil(t, refreshed.SSOAuthorizations)
	_, ok := refreshed.SSOAuthorizations[org.ID]
	require.True(t, ok)
}
