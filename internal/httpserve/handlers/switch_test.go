package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"

	ent "github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/privacy"
	"github.com/theopenlane/shared/enums"
	models "github.com/theopenlane/shared/openapi"
)

func (suite *HandlerTestSuite) TestSwitchHandlerSSOEnforced() {
	t := suite.T()

	// Create operation for SwitchHandler
	operation := suite.createImpersonationOperation("SwitchHandler", "Switch organization context")
	suite.registerTestHandler("POST", "switch", operation, suite.h.SwitchHandler)

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
		Name:      ulids.New().String(),
		SettingID: &setting.ID,
	}).SaveX(ownerCtx)

	suite.db.OrganizationSetting.UpdateOneID(setting.ID).SetOrganizationID(org.ID).ExecX(ownerCtx)

	// Test user attempting to switch.
	testUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      "0p3nl@n3rocks!",
		confirmedUser: true,
	})

	ctxTargetOrg := auth.NewTestContextWithOrgID(testUser.ID, org.ID)
	ctxTargetOrg = privacy.DecisionContext(ctxTargetOrg, privacy.Allow)
	testUserCtx := ent.NewContext(ctxTargetOrg, suite.db)

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

	// this is expected since the sso url will be generated
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestSwitchHandlerTFAEnforced() {
	t := suite.T()

	// Create operation for SwitchHandler
	operation := suite.createImpersonationOperation("SwitchHandler", "Switch organization context")
	suite.registerTestHandler("POST", "switch", operation, suite.h.SwitchHandler)

	ctx := echocontext.NewTestEchoContext().Request().Context()
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// Owner user to setup TFA organization.
	owner := suite.userBuilderWithInput(ctx, &userInput{
		password:      "0p3nl@n3rocks!",
		confirmedUser: true,
	})
	ownerCtx := privacy.DecisionContext(owner.UserCtx, privacy.Allow)
	ownerCtx = ent.NewContext(ownerCtx, suite.db)

	setting := suite.db.OrganizationSetting.Create().SetInput(ent.CreateOrganizationSettingInput{
		MultifactorAuthEnforced: func(b bool) *bool { return &b }(true),
	}).SaveX(ownerCtx)

	org := suite.db.Organization.Create().SetInput(ent.CreateOrganizationInput{
		Name:      ulids.New().String(),
		SettingID: &setting.ID,
	}).SaveX(ownerCtx)

	suite.db.OrganizationSetting.UpdateOneID(setting.ID).SetOrganizationID(org.ID).ExecX(ownerCtx)

	tfaenabled := false
	// Test user attempting to switch, without TFA enabled.
	testUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      "0p3nl@n3rocks!",
		confirmedUser: true,
		tfaEnabled:    tfaenabled, // User does not have TFA
	})

	ctxTargetOrg := auth.NewTestContextWithOrgID(testUser.ID, org.ID)
	ctxTargetOrg = privacy.DecisionContext(ctxTargetOrg, privacy.Allow)
	testUserCtx := ent.NewContext(ctxTargetOrg, suite.db)

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

	// Should return success with NeedsTFA flag
	assert.Equal(t, http.StatusOK, rec.Code)
	var out models.SwitchOrganizationReply
	err := json.NewDecoder(rec.Body).Decode(&out)
	assert.NoError(t, err)
	assert.True(t, out.Success)
	assert.True(t, out.NeedsTFA) // Should indicate TFA is needed
}

func (suite *HandlerTestSuite) TestSwitchHandlerTFAEnforcedUserHasTFA() {
	t := suite.T()

	// Create operation for SwitchHandler
	operation := suite.createImpersonationOperation("SwitchHandler", "Switch organization context")
	suite.registerTestHandler("POST", "switch", operation, suite.h.SwitchHandler)

	ctx := echocontext.NewTestEchoContext().Request().Context()
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// Owner user to setup TFA organization.
	owner := suite.userBuilderWithInput(ctx, &userInput{
		password:      "0p3nl@n3rocks!",
		confirmedUser: true,
	})
	ownerCtx := privacy.DecisionContext(owner.UserCtx, privacy.Allow)
	ownerCtx = ent.NewContext(ownerCtx, suite.db)

	setting := suite.db.OrganizationSetting.Create().SetInput(ent.CreateOrganizationSettingInput{
		MultifactorAuthEnforced: func(b bool) *bool { return &b }(true),
	}).SaveX(ownerCtx)

	org := suite.db.Organization.Create().SetInput(ent.CreateOrganizationInput{
		Name:      ulids.New().String(),
		SettingID: &setting.ID,
	}).SaveX(ownerCtx)

	suite.db.OrganizationSetting.UpdateOneID(setting.ID).SetOrganizationID(org.ID).ExecX(ownerCtx)

	tfaenabled := true
	// Test user attempting to switch, with TFA enabled.
	testUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      "0p3nl@n3rocks!",
		confirmedUser: true,
		tfaEnabled:    tfaenabled, // User has TFA
	})

	ctxTargetOrg := auth.NewTestContextWithOrgID(testUser.ID, org.ID)
	ctxTargetOrg = privacy.DecisionContext(ctxTargetOrg, privacy.Allow)
	testUserCtx := ent.NewContext(ctxTargetOrg, suite.db)

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

	// Should succeed normally since user has TFA
	assert.Equal(t, http.StatusOK, rec.Code)
	var out models.SwitchOrganizationReply
	err := json.NewDecoder(rec.Body).Decode(&out)
	assert.NoError(t, err)
	assert.True(t, out.Success)
	assert.False(t, out.NeedsTFA)       // Should not require TFA setup
	assert.NotEmpty(t, out.AccessToken) // Should have new auth data
}
