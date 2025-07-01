package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/httpsling"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
)

func (suite *HandlerTestSuite) TestSwitchHandlerSSOEnforced() {
	t := suite.T()

	suite.e.POST("switch", suite.h.SwitchHandler)

	ctx := echocontext.NewTestEchoContext().Request().Context()
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// Owner user to setup SSO organization.
	owner := suite.userBuilderWithInput(ctx, &userInput{
		password:      "0p3nl@n3rocks!",
		confirmedUser: true,
	})
	ownerCtx := privacy.DecisionContext(owner.UserCtx, privacy.Allow)
	ownerCtx = ent.NewContext(ownerCtx, suite.db)

	setting := suite.db.OrganizationSetting.Create().SetInput(ent.CreateOrganizationSettingInput{
		IdentityProviderLoginEnforced: func(b bool) *bool { return &b }(true),
	}).SaveX(ownerCtx)

	org := suite.db.Organization.Create().SetInput(ent.CreateOrganizationInput{
		Name:      gofakeit.Name(),
		SettingID: &setting.ID,
	}).SaveX(ownerCtx)

	suite.db.OrganizationSetting.UpdateOneID(setting.ID).SetOrganizationID(org.ID).ExecX(ownerCtx)

	// Test user attempting to switch.
	testUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      "0p3nl@n3rocks!",
		confirmedUser: true,
	})
	testUserCtx := privacy.DecisionContext(testUser.UserCtx, privacy.Allow)
	testUserCtx = ent.NewContext(testUserCtx, suite.db)

	suite.db.OrgMembership.Create().SetInput(ent.CreateOrgMembershipInput{
		OrganizationID: org.ID,
		UserID:         testUser.UserInfo.ID,
		Role:           &enums.RoleMember,
	}).ExecX(testUserCtx)

	suite.db.UserSetting.UpdateOneID(testUser.UserInfo.Edges.Setting.ID).SetDefaultOrgID(testUser.OrganizationID).ExecX(testUserCtx)

	body, _ := json.Marshal(models.SwitchOrganizationRequest{TargetOrganizationID: org.ID})

	req := httptest.NewRequest(http.MethodPost, "/switch", strings.NewReader(string(body)))
	req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	require.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "/v1/sso/login?organization_id="+org.ID, rec.Header().Get("Location"))
}
