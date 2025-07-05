package handlers_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ent "github.com/theopenlane/core/internal/ent/generated"
	generated "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/enums"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/echox/middleware/echocontext"
)

func ptr[T any](v T) *T { return &v }

func (suite *HandlerTestSuite) TestGoogleLoginHandlerSSOEnforced() {
	t := suite.T()

	login, _ := suite.h.GetGoogleLoginHandlers()
	suite.e.GET("google/login", echo.WrapHandler(login))

	ctx := echocontext.NewTestEchoContext().Request().Context()
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	ownerUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      "0wn3rP@ssw0rd",
		confirmedUser: true,
	})
	ownerCtx := privacy.DecisionContext(ownerUser.UserCtx, privacy.Allow)
	ownerCtx = ent.NewContext(ownerCtx, suite.db)

	setting := suite.db.OrganizationSetting.Create().SetInput(generated.CreateOrganizationSettingInput{
		IdentityProviderLoginEnforced: ptr(true),
	}).SaveX(ownerCtx)

	org := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      gofakeit.Name(),
		SettingID: &setting.ID,
	}).SaveX(ownerCtx)

	suite.db.OrganizationSetting.UpdateOneID(setting.ID).
		SetOrganizationID(org.ID).
		ExecX(ownerCtx)

	testUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      "$uper$ecretP@ssword",
		confirmedUser: true,
	})
	testUserCtx := privacy.DecisionContext(testUser.UserCtx, privacy.Allow)
	testUserCtx = ent.NewContext(testUserCtx, suite.db)

	suite.db.OrgMembership.Create().SetInput(generated.CreateOrgMembershipInput{
		OrganizationID: org.ID,
		UserID:         testUser.UserInfo.ID,
		Role:           &enums.RoleMember,
	}).ExecX(testUserCtx)

	suite.db.UserSetting.UpdateOneID(testUser.UserInfo.Edges.Setting.ID).SetDefaultOrgID(org.ID).ExecX(testUserCtx)

	req := httptest.NewRequest(http.MethodGet, "/google/login?email="+testUser.UserInfo.Email, nil)
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "/v1/sso/login?organization_id="+org.ID, rec.Header().Get("Location"))
}
func (suite *HandlerTestSuite) TestGoogleLoginHandlerSSOEnforcedOwnerBypass() {
	t := suite.T()

	login, _ := suite.h.GetGoogleLoginHandlers()
	suite.e.GET("google/login", echo.WrapHandler(login))

	ctx := echocontext.NewTestEchoContext().Request().Context()
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	ownerUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      "0wn3rP@ssw0rd",
		confirmedUser: true,
	})
	ownerCtx := privacy.DecisionContext(ownerUser.UserCtx, privacy.Allow)
	ownerCtx = ent.NewContext(ownerCtx, suite.db)

	setting := suite.db.OrganizationSetting.Create().SetInput(generated.CreateOrganizationSettingInput{
		IdentityProviderLoginEnforced: ptr(true),
	}).SaveX(ownerCtx)

	org := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      gofakeit.Name(),
		SettingID: &setting.ID,
	}).SaveX(ownerCtx)

	suite.db.OrganizationSetting.UpdateOneID(setting.ID).
		SetOrganizationID(org.ID).
		ExecX(ownerCtx)

	// user is set as owner by default
	suite.db.UserSetting.UpdateOneID(ownerUser.UserInfo.Edges.Setting.ID).SetDefaultOrgID(org.ID).ExecX(ownerCtx)

	req := httptest.NewRequest(http.MethodGet, "/google/login?email="+ownerUser.UserInfo.Email, nil)
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusFound, rec.Code)
	assert.NotContains(t, rec.Header().Get("Location"), "/v1/sso/login")
}
