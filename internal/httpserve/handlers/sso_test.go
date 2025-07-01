package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/ent/generated"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/usersetting"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
)

func (suite *HandlerTestSuite) TestWebfingerHandler() {
	t := suite.T()

	suite.e.GET(".well-known/webfinger", suite.h.WebfingerHandler)

	ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)
	ctx = ent.NewContext(ctx, suite.db)

	setting := suite.db.OrganizationSetting.Create().SetInput(generated.CreateOrganizationSettingInput{
		IdentityProviderLoginEnforced: lo.ToPtr(true),
		IdentityProvider:              lo.ToPtr(enums.SSOProviderOkta),
		OidcDiscoveryEndpoint:         lo.ToPtr("http://example.com"),
	}).SaveX(ctx)

	org := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      gofakeit.Name(),
		SettingID: &setting.ID,
	}).SaveX(ctx)

	suite.db.UserSetting.Update().Where(usersetting.UserID(testUser1.ID)).SetDefaultOrgID(org.ID).ExecX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/.well-known/webfinger?resource=org:"+org.ID, nil)
	rec := httptest.NewRecorder()

	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var out models.SSOStatusReply
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&out))
	log.Error().Err(errors.New("output")).Interface("out", out).Msg("WebfingerHandler output")
	assert.True(t, out.Enforced)
	assert.Equal(t, org.ID, out.OrganizationID)

	emailReq := httptest.NewRequest(http.MethodGet, "/.well-known/webfinger?resource=acct:"+testUser1.UserInfo.Email, nil)
	emailRec := httptest.NewRecorder()
	suite.e.ServeHTTP(emailRec, emailReq)
	require.Equal(t, http.StatusOK, emailRec.Code)
	var emailOut models.SSOStatusReply
	require.NoError(t, json.NewDecoder(emailRec.Body).Decode(&emailOut))
	assert.True(t, emailOut.Enforced)
	assert.Equal(t, org.ID, emailOut.OrganizationID)
}
