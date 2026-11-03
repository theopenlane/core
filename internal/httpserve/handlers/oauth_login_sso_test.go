package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	generated "github.com/theopenlane/core/v2/internal/ent/generated"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func ptr[T any](v T) *T { return &v }

func TestGoogleLoginHandlerSSOEnforced(t *testing.T) {
	suite := suite.setupParallelTest(t)

	login, _ := suite.h.GetGoogleLoginHandlers()
	suite.e.GET("google/login", echo.WrapHandler(login))

	ctx := echocontext.NewTestEchoContext().Request().Context()
	ctx = auth.WithInternalOperationContext(ctx)

	ownerUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      "0wn3rP@ssw0rd",
		confirmedUser: true,
	})
	ownerCtx := auth.WithInternalOperationContext(ownerUser.UserCtx)
	ownerCtx = ent.NewContext(ownerCtx, suite.db)

	setting, err := suite.db.OrganizationSetting.Create().Save(ownerCtx)
	assert.Check(t, err)

	org, err := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      ulids.New().String(),
		SettingID: &setting.ID,
	}).Save(ownerCtx)
	assert.Check(t, err)

	err = suite.db.OrganizationSetting.UpdateOneID(setting.ID).
		SetOrganizationID(org.ID).
		Exec(ownerCtx)
	assert.Check(t, err)

	suite.enforceSSOOnSetting(ownerCtx, setting.ID)

	testUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      "$uper$ecretP@ssword",
		confirmedUser: true,
	})

	ctxTargetOrg := auth.NewTestContextWithOrgID(testUser.ID, org.ID)
	ctxTargetOrg = auth.WithInternalOperationContext(ctxTargetOrg)
	testUserCtx := ent.NewContext(ctxTargetOrg, suite.db)

	err = suite.db.OrgMembership.Create().SetInput(generated.CreateOrgMembershipInput{
		OrganizationID: org.ID,
		UserID:         testUser.UserInfo.ID,
		Role:           &enums.RoleMember,
	}).Exec(testUserCtx)
	assert.Check(t, err)

	suite.db.UserSetting.UpdateOneID(testUser.UserInfo.Edges.Setting.ID).SetDefaultOrgID(org.ID).ExecX(testUserCtx)

	req := httptest.NewRequest(http.MethodGet, "/google/login?email="+testUser.UserInfo.Email, nil)
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	assert.Check(t, is.Equal(http.StatusFound, rec.Code))
	assert.Check(t, is.DeepEqual("/v1/sso/login?organization_id="+org.ID, rec.Header().Get("Location")))
}
func TestGoogleLoginHandlerSSOEnforcedOwnerBypass(t *testing.T) {
	suite := suite.setupParallelTest(t)

	login, _ := suite.h.GetGoogleLoginHandlers()
	suite.e.GET("google/login", echo.WrapHandler(login))

	ctx := echocontext.NewTestEchoContext().Request().Context()
	ctx = auth.WithInternalOperationContext(ctx)

	ownerUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      "0wn3rP@ssw0rd",
		confirmedUser: true,
	})
	ownerCtx := auth.WithInternalOperationContext(ownerUser.UserCtx)
	ownerCtx = ent.NewContext(ownerCtx, suite.db)

	setting, err := suite.db.OrganizationSetting.Create().Save(ownerCtx)
	assert.Check(t, err)

	org, err := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      ulids.New().String(),
		SettingID: &setting.ID,
	}).Save(ownerCtx)
	assert.Check(t, err)

	err = suite.db.OrganizationSetting.UpdateOneID(setting.ID).
		SetOrganizationID(org.ID).
		Exec(ownerCtx)
	assert.Check(t, err)

	suite.enforceSSOOnSetting(ownerCtx, setting.ID)

	// user is set as owner by default
	suite.db.UserSetting.UpdateOneID(ownerUser.UserInfo.Edges.Setting.ID).SetDefaultOrgID(org.ID).ExecX(ownerCtx)

	req := httptest.NewRequest(http.MethodGet, "/google/login?email="+ownerUser.UserInfo.Email, nil)
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	assert.Check(t, is.Equal(http.StatusFound, rec.Code))
	assert.Check(t, !strings.Contains(rec.Header().Get("Location"), "/v1/sso/login"))
}

func TestGoogleLoginHandlerTFAEnforced(t *testing.T) {
	suite := suite.setupParallelTest(t)

	login, _ := suite.h.GetGoogleLoginHandlers()
	suite.e.GET("google/login", echo.WrapHandler(login))

	ctx := echocontext.NewTestEchoContext().Request().Context()
	ctx = auth.WithInternalOperationContext(ctx)

	ownerUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      "0wn3rP@ssw0rd",
		confirmedUser: true,
	})
	ownerCtx := auth.WithInternalOperationContext(ownerUser.UserCtx)
	ownerCtx = ent.NewContext(ownerCtx, suite.db)

	// Create org with TFA enforced
	setting, err := suite.db.OrganizationSetting.Create().SetInput(generated.CreateOrganizationSettingInput{
		MultifactorAuthEnforced: ptr(true),
	}).Save(ownerCtx)
	assert.Check(t, err)

	org, err := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      ulids.New().String(),
		SettingID: &setting.ID,
	}).Save(ownerCtx)
	assert.Check(t, err)

	err = suite.db.OrganizationSetting.UpdateOneID(setting.ID).
		SetOrganizationID(org.ID).
		Exec(ownerCtx)
	assert.Check(t, err)

	tfaenabled := false
	testUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      "$uper$ecretP@ssword",
		confirmedUser: true,
		tfaEnabled:    tfaenabled, // User without TFA
	})

	ctxTargetOrg := auth.NewTestContextWithOrgID(testUser.ID, org.ID)
	ctxTargetOrg = auth.WithInternalOperationContext(ctxTargetOrg)
	testUserCtx := ent.NewContext(ctxTargetOrg, suite.db)

	err = suite.db.OrgMembership.Create().SetInput(generated.CreateOrgMembershipInput{
		OrganizationID: org.ID,
		UserID:         testUser.UserInfo.ID,
		Role:           &enums.RoleMember,
	}).Exec(testUserCtx)
	assert.Check(t, err)

	suite.db.UserSetting.UpdateOneID(testUser.UserInfo.Edges.Setting.ID).SetDefaultOrgID(org.ID).ExecX(testUserCtx)

	// Note: Currently OAuth login handlers don't check for TFA, they proceed normally
	// This test documents current behavior - TFA check happens at session/token generation time
	req := httptest.NewRequest(http.MethodGet, "/google/login?email="+testUser.UserInfo.Email, nil)
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	assert.Check(t, is.Equal(http.StatusFound, rec.Code))
	// OAuth login proceeds normally - TFA would be enforced after successful OAuth authentication
	assert.Check(t, is.Contains(rec.Header().Get("Location"), "accounts.google.com"))
}

func TestGithubLoginHandlerTFAEnforced(t *testing.T) {
	suite := suite.setupParallelTest(t)

	login, _ := suite.h.GetGitHubLoginHandlers()
	suite.e.GET("github/login", echo.WrapHandler(login))

	ctx := echocontext.NewTestEchoContext().Request().Context()
	ctx = auth.WithInternalOperationContext(ctx)

	ownerUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      "0wn3rP@ssw0rd",
		confirmedUser: true,
	})
	ownerCtx := auth.WithInternalOperationContext(ownerUser.UserCtx)
	ownerCtx = ent.NewContext(ownerCtx, suite.db)

	// Create org with TFA enforced
	setting, err := suite.db.OrganizationSetting.Create().SetInput(generated.CreateOrganizationSettingInput{
		MultifactorAuthEnforced: ptr(true),
	}).Save(ownerCtx)
	assert.Check(t, err)

	org, err := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      ulids.New().String(),
		SettingID: &setting.ID,
	}).Save(ownerCtx)
	assert.Check(t, err)

	err = suite.db.OrganizationSetting.UpdateOneID(setting.ID).
		SetOrganizationID(org.ID).
		Exec(ownerCtx)
	assert.Check(t, err)

	tfaenabled := false
	testUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      "$uper$ecretP@ssword",
		confirmedUser: true,
		tfaEnabled:    tfaenabled, // User without TFA
	})

	ctxTargetOrg := auth.NewTestContextWithOrgID(testUser.ID, org.ID)
	ctxTargetOrg = auth.WithInternalOperationContext(ctxTargetOrg)
	testUserCtx := ent.NewContext(ctxTargetOrg, suite.db)

	err = suite.db.OrgMembership.Create().SetInput(generated.CreateOrgMembershipInput{
		OrganizationID: org.ID,
		UserID:         testUser.UserInfo.ID,
		Role:           &enums.RoleMember,
	}).Exec(testUserCtx)
	assert.Check(t, err)

	suite.db.UserSetting.UpdateOneID(testUser.UserInfo.Edges.Setting.ID).SetDefaultOrgID(org.ID).ExecX(testUserCtx)

	// Note: Currently OAuth login handlers don't check for TFA, they proceed normally
	// This test documents current behavior - TFA check happens at session/token generation time
	req := httptest.NewRequest(http.MethodGet, "/github/login?email="+testUser.UserInfo.Email, nil)
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	assert.Check(t, is.Equal(http.StatusFound, rec.Code))
	// OAuth login proceeds normally - TFA would be enforced after successful OAuth authentication
	assert.Check(t, is.Contains(rec.Header().Get("Location"), "github.com/login/oauth/authorize"))
}
