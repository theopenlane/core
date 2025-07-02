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

	suite.e.GET("sso/token/authorize", suite.h.SSOTokenAuthorizeHandler)

	ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)
	ctx = ent.NewContext(ctx, suite.db)

	setting := suite.db.OrganizationSetting.Create().SetInput(ent.CreateOrganizationSettingInput{
		IdentityProviderLoginEnforced: lo.ToPtr(true),
		IdentityProvider:              lo.ToPtr(enums.SSOProviderOkta),
		OidcDiscoveryEndpoint:         lo.ToPtr("http://example.com"),
	}).SaveX(ctx)

	org := suite.db.Organization.Create().SetInput(ent.CreateOrganizationInput{
		Name:      gofakeit.Name(),
		SettingID: &setting.ID,
	}).SaveX(ctx)

	token := suite.db.APIToken.Create().SetOwnerID(org.ID).SetName("t").SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, sso.SSOTokenAuthorize(nil, org.ID, token.ID, "api"), nil)
	rec := httptest.NewRecorder()

	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusFound, rec.Code)
}
