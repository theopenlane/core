package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/samber/lo"

	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/v2/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/v2/internal/ent/generated/usersetting"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

// enforceSSOOnSetting configures a tested identity provider on the organization setting and turns on SSO
// login enforcement, following the configure -> mark tested -> enforce sequence the organization setting
// hooks require before enforcement is allowed
func (suite *HandlerTestSuite) enforceSSOOnSetting(ctx context.Context, settingID string) {
	t := suite.T()

	assert.NilError(t, suite.db.OrganizationSetting.UpdateOneID(settingID).
		SetIdentityProvider(enums.SSOProviderOkta).
		SetIdentityProviderClientID("client").
		SetIdentityProviderClientSecret("secret").
		SetOidcDiscoveryEndpoint("http://example.com").
		Exec(ctx))

	assert.NilError(t, suite.db.OrganizationSetting.UpdateOneID(settingID).
		SetIdentityProviderAuthTested(true).
		Exec(ctx))

	assert.NilError(t, suite.db.OrganizationSetting.UpdateOneID(settingID).
		SetIdentityProviderLoginEnforced(true).
		Exec(ctx))
}

// ssoEnforcedOrg creates an SSO-enforced organization owned by a fresh user and returns it
func (suite *HandlerTestSuite) ssoEnforcedOrg() *ent.Organization {
	ctx := echocontext.NewTestEchoContext().Request().Context()

	owner := suite.userBuilderWithInput(ctx, &userInput{password: "0wn3rP@ssw0rd!", confirmedUser: true})
	ownerCtx := auth.WithInternalOperationContext(owner.UserCtx)
	ownerCtx = ent.NewContext(ownerCtx, suite.db)

	t := suite.T()

	setting, err := suite.db.OrganizationSetting.Create().Save(ownerCtx)
	assert.NilError(t, err)

	org, err := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      ulids.New().String(),
		SettingID: &setting.ID,
	}).Save(ownerCtx)
	assert.NilError(t, err)

	assert.NilError(t, suite.db.OrganizationSetting.UpdateOneID(setting.ID).SetOrganizationID(org.ID).Exec(ownerCtx))

	suite.enforceSSOOnSetting(ownerCtx, setting.ID)

	return org
}

func TestLoginHandlerSSOExemptMember(t *testing.T) {
	suite.setupTest(t)

	suite.registerTestHandler("POST", "login", suite.h.LoginHandler)

	org := suite.ssoEnforcedOrg()

	ctx := echocontext.NewTestEchoContext().Request().Context()
	member := suite.userBuilderWithInput(ctx, &userInput{password: "$uper$ecretP@ssword", confirmedUser: true})

	memberCtx := auth.NewTestContextWithOrgID(member.ID, org.ID)
	memberCtx = auth.WithInternalOperationContext(memberCtx)
	memberCtx = ent.NewContext(memberCtx, suite.db)

	// member is explicitly exempt from SSO for this organization
	assert.NilError(t, suite.db.OrgMembership.Create().SetInput(generated.CreateOrgMembershipInput{
		OrganizationID: org.ID,
		UserID:         member.UserInfo.ID,
		Role:           &enums.RoleMember,
		SSOExempt:      lo.ToPtr(true),
	}).Exec(memberCtx))

	assert.NilError(t, suite.db.UserSetting.UpdateOneID(member.UserInfo.Edges.Setting.ID).SetDefaultOrgID(org.ID).Exec(memberCtx))

	body, _ := json.Marshal(models.LoginRequest{Username: member.UserInfo.Email, Password: "$uper$ecretP@ssword"})
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(string(body)))
	req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	// the exempt member is not redirected to SSO; password login succeeds
	assert.Assert(t, is.Equal(http.StatusOK, rec.Code))
	var out models.LoginResponse
	assert.NilError(t, json.NewDecoder(rec.Body).Decode(&out))
	assert.Check(t, out.Success)
}

func TestLoginHandlerSSOExemptDomain(t *testing.T) {
	suite.setupTest(t)

	suite.registerTestHandler("POST", "login", suite.h.LoginHandler)

	org := suite.ssoEnforcedOrg()

	ctx := auth.WithInternalCrossOrgContext(echocontext.NewTestEchoContext().Request().Context())
	ctx = ent.NewContext(ctx, suite.db)

	exemptDomain := strings.ToLower(ulids.New().String()) + ".example.com"
	assert.NilError(t, suite.db.OrganizationSetting.Update().
		Where(organizationsetting.OrganizationID(org.ID)).
		SetSSOExemptDomains([]string{exemptDomain}).
		Exec(ctx))

	memberEmail := "auditor" + strings.ToLower(ulids.New().String()) + "@" + exemptDomain
	member := suite.userBuilderWithInput(ctx, &userInput{email: memberEmail, password: "$uper$ecretP@ssword", confirmedUser: true})

	memberCtx := auth.NewTestContextWithOrgID(member.ID, org.ID)
	memberCtx = auth.WithInternalOperationContext(memberCtx)
	memberCtx = ent.NewContext(memberCtx, suite.db)

	// member relies on the per-domain exemption, not a per-user flag
	assert.NilError(t, suite.db.OrgMembership.Create().SetInput(generated.CreateOrgMembershipInput{
		OrganizationID: org.ID,
		UserID:         member.UserInfo.ID,
		Role:           &enums.RoleMember,
	}).Exec(memberCtx))

	assert.NilError(t, suite.db.UserSetting.UpdateOneID(member.UserInfo.Edges.Setting.ID).SetDefaultOrgID(org.ID).Exec(memberCtx))

	body, _ := json.Marshal(models.LoginRequest{Username: member.UserInfo.Email, Password: "$uper$ecretP@ssword"})
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(string(body)))
	req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	assert.Assert(t, is.Equal(http.StatusOK, rec.Code))
	var out models.LoginResponse
	assert.NilError(t, json.NewDecoder(rec.Body).Decode(&out))
	assert.Check(t, out.Success)
}

func TestWebfingerHandlerExemptMember(t *testing.T) {
	suite.setupTest(t)

	suite.registerTestHandler("GET", ".well-known/webfinger", suite.h.WebfingerHandler)

	org := suite.ssoEnforcedOrg()

	ctx := echocontext.NewTestEchoContext().Request().Context()
	member := suite.userBuilderWithInput(ctx, &userInput{confirmedUser: true})

	memberCtx := auth.NewTestContextWithOrgID(member.ID, org.ID)
	memberCtx = auth.WithInternalOperationContext(memberCtx)
	memberCtx = ent.NewContext(memberCtx, suite.db)

	assert.NilError(t, suite.db.OrgMembership.Create().SetInput(generated.CreateOrgMembershipInput{
		OrganizationID: org.ID,
		UserID:         member.UserInfo.ID,
		Role:           &enums.RoleMember,
		SSOExempt:      lo.ToPtr(true),
	}).Exec(memberCtx))

	assert.NilError(t, suite.db.UserSetting.Update().Where(usersetting.UserID(member.ID)).SetDefaultOrgID(org.ID).Exec(memberCtx))

	// the org level lookup still reports enforcement
	orgReq := httptest.NewRequest(http.MethodGet, "/.well-known/webfinger?resource=org:"+org.ID, nil)
	orgRec := httptest.NewRecorder()
	suite.e.ServeHTTP(orgRec, orgReq)
	assert.Assert(t, is.Equal(http.StatusOK, orgRec.Code))
	var orgOut models.SSOStatusResponse
	assert.NilError(t, json.NewDecoder(orgRec.Body).Decode(&orgOut))
	assert.Check(t, orgOut.Enforced, "org level lookup reports enforcement")

	// the account lookup reflects the member's exemption
	acctReq := httptest.NewRequest(http.MethodGet, "/.well-known/webfinger?resource=acct:"+member.UserInfo.Email, nil)
	acctRec := httptest.NewRecorder()
	suite.e.ServeHTTP(acctRec, acctReq)
	assert.Assert(t, is.Equal(http.StatusOK, acctRec.Code))
	var acctOut models.SSOStatusResponse
	assert.NilError(t, json.NewDecoder(acctRec.Body).Decode(&acctOut))
	assert.Check(t, !acctOut.Enforced, "exempt member is not required to use SSO")
}

func TestOrgOwnerSeededSSOExempt(t *testing.T) {
	suite.setupTest(t)

	ctx := auth.WithInternalOperationContext(testUser1.UserCtx)
	ctx = ent.NewContext(ctx, suite.db)

	org, err := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name: ulids.New().String(),
	}).Save(ctx)
	assert.NilError(t, err)

	member, err := suite.db.OrgMembership.Query().
		Where(orgmembership.OrganizationID(org.ID), orgmembership.UserID(testUser1.ID)).
		Only(ctx)
	assert.NilError(t, err)

	assert.Check(t, is.DeepEqual(enums.RoleOwner, member.Role))
	assert.Check(t, member.SSOExempt, "organization owner is auto-seeded as SSO exempt")
	assert.Assert(t, member.SSOExemptReason != nil)
	assert.Check(t, is.Equal("organization owner", *member.SSOExemptReason))
}

func TestSwitchHandlerSSOExemptMember(t *testing.T) {
	suite.setupTest(t)

	suite.registerTestHandler("POST", "switch", suite.h.SwitchHandler)

	org := suite.ssoEnforcedOrg()

	ctx := echocontext.NewTestEchoContext().Request().Context()
	member := suite.userBuilderWithInput(ctx, &userInput{password: "0p3nl@n3rocks!", confirmedUser: true})

	memberCtx := auth.NewTestContextWithOrgID(member.ID, org.ID)
	memberCtx = auth.WithInternalOperationContext(memberCtx)
	memberCtx = ent.NewContext(memberCtx, suite.db)

	assert.NilError(t, suite.db.OrgMembership.Create().SetInput(generated.CreateOrgMembershipInput{
		OrganizationID: org.ID,
		UserID:         member.UserInfo.ID,
		Role:           &enums.RoleMember,
		SSOExempt:      lo.ToPtr(true),
	}).Exec(memberCtx))

	body, _ := json.Marshal(models.SwitchOrganizationRequest{TargetOrganizationID: org.ID})
	req := httptest.NewRequest(http.MethodPost, "/switch", strings.NewReader(string(body)))
	req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(member.UserCtx))

	// the exempt member switches into the SSO-enforced org without being redirected through SSO
	assert.Assert(t, is.Equal(http.StatusOK, rec.Code))
	var out models.SwitchOrganizationResponse
	assert.NilError(t, json.NewDecoder(rec.Body).Decode(&out))
	assert.Check(t, !out.NeedsSSO, "exempt member is not required to use SSO when switching")
}

func TestInviteGrantsSSOExempt(t *testing.T) {
	suite.setupTest(t)

	ctx := auth.WithInternalOperationContext(testUser1.UserCtx)
	ctx = ent.NewContext(ctx, suite.db)

	orgID := testUser1.OrganizationID

	recipientEmail := "auditor" + strings.ToLower(ulids.New().String()) + "@partner.example.com"
	recipient, err := suite.db.User.Create().
		SetEmail(recipientEmail).
		SetFirstName("Audit").
		SetLastName("Or").
		SetAuthProvider(enums.AuthProviderCredentials).
		SetLastLoginProvider(enums.AuthProviderCredentials).
		SetLastSeen(time.Now()).
		Save(ctx)
	assert.NilError(t, err)

	// an invitation that grants an SSO exemption on acceptance
	role := enums.RoleMember
	inv, err := suite.db.Invite.Create().
		SetRecipient(recipientEmail).
		SetRole(role).
		SetSSOExempt(true).
		Save(ctx)
	assert.NilError(t, err)

	// accepting the invite as the recipient triggers the accepted hook, which creates the membership
	recipientCtx := auth.NewTestContextWithOrgID(recipient.ID, orgID)
	recipientCtx = auth.WithInternalOperationContext(recipientCtx)
	recipientCtx = ent.NewContext(recipientCtx, suite.db)

	assert.NilError(t, suite.db.Invite.UpdateOneID(inv.ID).SetStatus(enums.InvitationAccepted).Exec(recipientCtx))

	member, err := suite.db.OrgMembership.Query().
		Where(orgmembership.OrganizationID(orgID), orgmembership.UserID(recipient.ID)).
		Only(ctx)
	assert.NilError(t, err)

	assert.Check(t, member.SSOExempt, "invite-granted exemption sets the membership exemption")
	assert.Assert(t, member.SSOExemptReason != nil)
	assert.Check(t, is.Equal("granted via invitation", *member.SSOExemptReason))
}

// TestImpersonatorAttributionStamped verifies that records mutated during an impersonation session
// record the acting individual in updated_by_impersonator, while the standard audit fields continue to
// reflect the impersonated identity
func TestImpersonatorAttributionStamped(t *testing.T) {
	suite.setupTest(t)

	base := testUser1.UserCtx

	caller := &auth.Caller{
		SubjectID:       testUser1.ID,
		SubjectEmail:    testUser1.UserInfo.Email,
		OrganizationID:  testUser1.OrganizationID,
		OrganizationIDs: []string{testUser1.OrganizationID},
		Impersonation: &auth.ImpersonationContext{
			Type:              auth.SupportImpersonation,
			ImpersonatorID:    "engineer@theopenlane.io",
			ImpersonatorEmail: "engineer@theopenlane.io",
		},
	}

	impCtx := auth.WithCaller(base, caller)
	impCtx = auth.WithInternalOperationContext(impCtx)
	impCtx = ent.NewContext(impCtx, suite.db)

	group, err := suite.db.Group.Create().
		SetName("Support Touched " + ulids.New().String()).
		SetOwnerID(testUser1.OrganizationID).
		Save(impCtx)
	assert.NilError(t, err)

	assert.Assert(t, group.UpdatedByImpersonator != nil)
	assert.Check(t, is.Equal("engineer@theopenlane.io", *group.UpdatedByImpersonator))
}

// TestExemptDomainAllowedDomainOverlap verifies a domain may appear in both the sso exempt domains and the
// allowed email domains lists; the two are independent because allowed email domains only govern auto-join
// when SSO is not enforced, while exempt domains only affect the SSO redirect when it is enforced
func TestExemptDomainAllowedDomainOverlap(t *testing.T) {
	suite.setupTest(t)

	ctx := echocontext.NewTestEchoContext().Request().Context()
	ctx = auth.WithInternalOperationContext(ctx)
	ctx = ent.NewContext(ctx, suite.db)

	setting, err := suite.db.OrganizationSetting.Create().
		SetAllowedEmailDomains([]string{"audit.com"}).
		Save(ctx)
	assert.NilError(t, err)

	// the same domain may also be marked sso exempt; previously this was rejected as a conflict
	updated, err := suite.db.OrganizationSetting.UpdateOneID(setting.ID).
		SetSSOExemptDomains([]string{"audit.com"}).
		Save(ctx)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual([]string{"audit.com"}, updated.AllowedEmailDomains))
	assert.Check(t, is.DeepEqual([]string{"audit.com"}, updated.SSOExemptDomains))
}

// TestAutoJoinSuppressedWhenSSOEnforced verifies that domain-based auto-join does not add a user to an
// organization that enforces SSO; for enforced organizations membership comes through the identity
// provider (JIT) instead
func TestAutoJoinSuppressedWhenSSOEnforced(t *testing.T) {
	suite.setupTest(t)

	// a real caller is required so the organization create hook can assign the owner membership
	ctx := auth.WithInternalOperationContext(testUser1.UserCtx)
	ctx = ent.NewContext(ctx, suite.db)

	const allowedDomain = "enforced-autojoin.com"

	// an organization that has domain auto-join configured AND enforces SSO
	setting, err := suite.db.OrganizationSetting.Create().SetInput(generated.CreateOrganizationSettingInput{
		AllowedEmailDomains:          []string{allowedDomain},
		AllowMatchingDomainsAutojoin: lo.ToPtr(true),
	}).Save(ctx)
	assert.NilError(t, err)

	org, err := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      ulids.New().String(),
		SettingID: &setting.ID,
	}).Save(ctx)
	assert.NilError(t, err)

	assert.NilError(t, suite.db.OrganizationSetting.UpdateOneID(setting.ID).SetOrganizationID(org.ID).Exec(ctx))
	suite.enforceSSOOnSetting(ctx, setting.ID)

	// a brand-new user on the allowed domain confirms their email, which triggers the auto-join hook
	userSetting, err := suite.db.UserSetting.Create().SetEmailConfirmed(false).Save(ctx)
	assert.NilError(t, err)

	user, err := suite.db.User.Create().
		SetFirstName("Auto").
		SetLastName("Join").
		SetEmail(ulids.New().String() + "@" + allowedDomain).
		SetSetting(userSetting).
		SetLastLoginProvider(enums.AuthProviderCredentials).
		SetLastSeen(time.Now()).
		Save(ctx)
	assert.NilError(t, err)

	assert.NilError(t, suite.db.UserSetting.UpdateOneID(userSetting.ID).SetEmailConfirmed(true).Exec(ctx))

	// the enforced organization must not have auto-joined the user
	exists, err := suite.db.OrgMembership.Query().
		Where(orgmembership.UserID(user.ID), orgmembership.OrganizationID(org.ID)).
		Exist(ctx)
	assert.NilError(t, err)
	assert.Check(t, !exists, "auto-join must be suppressed for SSO-enforced organizations")
}
