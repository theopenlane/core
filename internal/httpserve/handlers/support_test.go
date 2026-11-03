package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/samber/lo"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"github.com/theopenlane/core/v2/internal/httpserve/handlers"
	"github.com/theopenlane/core/v2/pkg/middleware/impersonation"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

// supportTestConfig returns a SupportAccessConfig pointed at the given mock OIDC server, with a known
// password and the theopenlane.io domain restriction
func supportTestConfig(issuer string) handlers.SupportAccessConfig {
	return handlers.SupportAccessConfig{
		Enabled:           true,
		Email:             "support@theopenlane.io",
		DisplayName:       "Openlane Support",
		SubjectID:         auth.SupportSubjectID,
		Password:          "super-secret-support-password",
		ClientID:          "support-client",
		ClientSecret:      "secret",
		IssuerURL:         issuer,
		DiscoveryEndpoint: issuer + "/.well-known/openid-configuration",
		AllowedDomain:     "theopenlane.io",
	}
}

// createConsentingOrg creates an organization whose setting consents to support access
func (suite *HandlerTestSuite) createConsentingOrg(ctx context.Context) *ent.Organization {
	t := suite.T()

	setting, err := suite.db.OrganizationSetting.Create().SetInput(generated.CreateOrganizationSettingInput{
		AllowSupportAccess: lo.ToPtr(true),
	}).Save(ctx)
	assert.NilError(t, err)

	org, err := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      ulids.New().String(),
		SettingID: &setting.ID,
	}).Save(ctx)
	assert.NilError(t, err)

	assert.NilError(t, suite.db.OrganizationSetting.UpdateOneID(setting.ID).SetOrganizationID(org.ID).Exec(ctx))

	return org
}

func TestSupportAccessLoginAndCallback(t *testing.T) {
	suite := suite.setupTest(t)

	suite.registerTestHandler("POST", "v1/login", suite.h.LoginHandler)
	suite.registerTestHandler("POST", "v1/support/callback", suite.h.SupportCallbackHandler)

	testCases := []struct {
		name           string
		reason         string
		expectedReason string
	}{
		{
			name:           "plain reason",
			reason:         "assisting customer with data import",
			expectedReason: "assisting customer with data import",
		},
		{
			name:           "url escaped reason",
			reason:         url.QueryEscape("this is an escaped text"),
			expectedReason: "this is an escaped text",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			oidc := newMockOIDCServer(t,
				withExpectedCode("code123"),
				withClientSecret("secret"),
				withUserInfo(suite.testUser1.UserInfo.Email, "Support Engineer", ""),
			)
			defer oidc.Close()

			// configure support access on the handler and restore afterwards so other tests are unaffected
			original := suite.h.SupportAccessConfig
			suite.h.SupportAccessConfig = supportTestConfig(oidc.server.URL)
			suite.h.SupportAccessConfig.AllowedDomain = strings.SplitN(suite.testUser1.UserInfo.Email, "@", 2)[1]
			defer func() { suite.h.SupportAccessConfig = original }()

			ctx := auth.WithInternalOperationContext(suite.testUser1.UserCtx)
			ctx = ent.NewContext(ctx, suite.db)

			org := suite.createConsentingOrg(ctx)

			// first factor: authenticate the support identity against the configured password
			loginBody, _ := json.Marshal(models.LoginRequest{
				Username:             "support@theopenlane.io",
				Password:             "super-secret-support-password",
				TargetOrganizationID: org.ID,
				Reason:               tc.reason,
			})

			loginReq := httptest.NewRequest(http.MethodPost, "/v1/login", strings.NewReader(string(loginBody)))
			loginReq.Header.Set("Content-Type", "application/json")
			loginRec := httptest.NewRecorder()
			suite.e.ServeHTTP(loginRec, loginReq)

			assert.Assert(t, is.Equal(http.StatusOK, loginRec.Code))

			var loginOut models.LoginResponse
			assert.NilError(t, json.NewDecoder(loginRec.Body).Decode(&loginOut))
			assert.Check(t, loginOut.Success)
			assert.Check(t, len(loginOut.RedirectURI) > 0, "first factor should return the identity provider redirect")

			// collect the support cookies set by the first factor
			cookieParts := []string{}
			var state string
			for _, c := range loginRec.Result().Cookies() {
				cookieParts = append(cookieParts, c.Name+"="+c.Value)
				if c.Name == "support_state" {
					state = c.Value
				}
				if c.Name == "support_nonce" {
					oidc.nonce = c.Value
				}
			}

			assert.Assert(t, len(state) > 0, "support_state cookie should be set")

			// second factor: complete the identity provider exchange
			cbBody, _ := json.Marshal(models.SupportCallbackRequest{Code: "code123", State: state})
			cbReq := httptest.NewRequest(http.MethodPost, "/v1/support/callback", strings.NewReader(string(cbBody)))
			cbReq.Header.Set("Content-Type", "application/json")
			cbReq.Header.Set("Cookie", strings.Join(cookieParts, "; "))
			cbRec := httptest.NewRecorder()
			suite.e.ServeHTTP(cbRec, cbReq)

			assert.Assert(t, is.Equal(http.StatusOK, cbRec.Code))

			var cbOut models.SupportAccessResponse
			assert.NilError(t, json.NewDecoder(cbRec.Body).Decode(&cbOut))
			assert.Check(t, cbOut.Success)
			assert.Check(t, len(cbOut.Token) > 0)
			assert.Check(t, is.DeepEqual(org.ID, cbOut.OrganizationID))
			assert.Check(t, is.DeepEqual(suite.testUser1.UserInfo.Email, cbOut.Impersonator))

			// the minted token must carry both identities: the virtual support user and the individual
			claims, err := suite.h.TokenManager.ValidateImpersonationToken(context.Background(), cbOut.Token)
			assert.NilError(t, err)
			assert.Check(t, is.DeepEqual(auth.SupportSubjectID, claims.UserID), "target is the virtual support identity")
			assert.Check(t, is.DeepEqual(suite.testUser1.ID, claims.ImpersonatorID), "impersonator is the individual from the IdP")
			assert.Check(t, is.Equal("support", claims.Type))
			assert.Check(t, is.DeepEqual(org.ID, claims.OrgID))
			assert.Check(t, is.DeepEqual(tc.expectedReason, claims.Reason))
		})
	}
}

func TestSupportAccessRejectsWrongPassword(t *testing.T) {
	suite := suite.setupTest(t)

	suite.registerTestHandler("POST", "v1/login", suite.h.LoginHandler)

	oidc := newMockOIDCServer(t)
	defer oidc.Close()

	original := suite.h.SupportAccessConfig
	suite.h.SupportAccessConfig = supportTestConfig(oidc.server.URL)
	defer func() { suite.h.SupportAccessConfig = original }()

	ctx := auth.WithInternalOperationContext(suite.testUser1.UserCtx)
	ctx = ent.NewContext(ctx, suite.db)
	org := suite.createConsentingOrg(ctx)

	body, _ := json.Marshal(models.LoginRequest{
		Username:             "support@theopenlane.io",
		Password:             "wrong-password",
		TargetOrganizationID: org.ID,
		Reason:               "assisting customer with data import",
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/login", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	assert.Check(t, is.Equal(http.StatusUnauthorized, rec.Code))
}

func TestSupportAccessRejectsNonConsentingOrg(t *testing.T) {
	suite := suite.setupTest(t)

	suite.registerTestHandler("POST", "v1/login", suite.h.LoginHandler)

	oidc := newMockOIDCServer(t)
	defer oidc.Close()

	original := suite.h.SupportAccessConfig
	suite.h.SupportAccessConfig = supportTestConfig(oidc.server.URL)
	defer func() { suite.h.SupportAccessConfig = original }()

	ctx := auth.WithInternalOperationContext(suite.testUser1.UserCtx)
	ctx = ent.NewContext(ctx, suite.db)

	// org that has NOT consented to support access
	setting, err := suite.db.OrganizationSetting.Create().SetInput(generated.CreateOrganizationSettingInput{}).Save(ctx)
	assert.NilError(t, err)

	org, err := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      ulids.New().String(),
		SettingID: &setting.ID,
	}).Save(ctx)
	assert.NilError(t, err)

	assert.NilError(t, suite.db.OrganizationSetting.UpdateOneID(setting.ID).SetOrganizationID(org.ID).Exec(ctx))

	body, _ := json.Marshal(models.LoginRequest{
		Username:             "support@theopenlane.io",
		Password:             "super-secret-support-password",
		TargetOrganizationID: org.ID,
		Reason:               "assisting customer with data import",
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/login", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	assert.Check(t, is.Equal(http.StatusForbidden, rec.Code))
}

// TestSupportAccessRejectsDomainMismatch ensures the second factor rejects an individual outside the
// configured domain even when the first factor and consent succeed
func TestSupportAccessRejectsDomainMismatch(t *testing.T) {
	suite := suite.setupTest(t)

	suite.registerTestHandler("POST", "v1/login", suite.h.LoginHandler)
	suite.registerTestHandler("POST", "v1/support/callback", suite.h.SupportCallbackHandler)

	// individual from a domain that is not allowed
	oidc := newMockOIDCServer(t,
		withExpectedCode("code123"),
		withClientSecret("secret"),
		withUserInfo("attacker@evil.com", "Not Allowed", ""),
	)
	defer oidc.Close()

	original := suite.h.SupportAccessConfig
	suite.h.SupportAccessConfig = supportTestConfig(oidc.server.URL)
	defer func() { suite.h.SupportAccessConfig = original }()

	ctx := auth.WithInternalOperationContext(suite.testUser1.UserCtx)
	ctx = ent.NewContext(ctx, suite.db)
	org := suite.createConsentingOrg(ctx)

	loginBody, _ := json.Marshal(models.LoginRequest{
		Username:             "support@theopenlane.io",
		Password:             "super-secret-support-password",
		TargetOrganizationID: org.ID,
		Reason:               "assisting customer with data import",
	})
	loginReq := httptest.NewRequest(http.MethodPost, "/v1/login", strings.NewReader(string(loginBody)))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	suite.e.ServeHTTP(loginRec, loginReq)
	assert.Assert(t, is.Equal(http.StatusOK, loginRec.Code))

	cookieParts := []string{}
	var state string
	for _, c := range loginRec.Result().Cookies() {
		cookieParts = append(cookieParts, c.Name+"="+c.Value)
		if c.Name == "support_state" {
			state = c.Value
		}
		if c.Name == "support_nonce" {
			oidc.nonce = c.Value
		}
	}

	cbBody, _ := json.Marshal(models.SupportCallbackRequest{Code: "code123", State: state})
	cbReq := httptest.NewRequest(http.MethodPost, "/v1/support/callback", strings.NewReader(string(cbBody)))
	cbReq.Header.Set("Content-Type", "application/json")
	cbReq.Header.Set("Cookie", strings.Join(cookieParts, "; "))
	cbRec := httptest.NewRecorder()
	suite.e.ServeHTTP(cbRec, cbReq)

	assert.Check(t, is.Equal(http.StatusForbidden, cbRec.Code))
}

// supportCallerContext resolves an issued support token through the impersonation middleware exactly as
// the router does, returning the request context carrying the virtual support caller
func (suite *HandlerTestSuite) supportCallerContext(token string) context.Context {
	mw := impersonation.New(suite.h.TokenManager, suite.h.SupportAccessConfig.SubjectID, suite.h.SupportAccessConfig.DisplayName)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(auth.Authorization, auth.ImpersonationScheme+" "+token)

	c := suite.e.NewContext(req, httptest.NewRecorder())

	var resolved context.Context

	err := mw.Process(func(c echo.Context) error {
		resolved = c.Request().Context()
		return nil
	})(c)
	assert.NilError(suite.T(), err)

	return ent.NewContext(resolved, suite.db)
}

// supportSessionToken runs the full two-factor support login against a mock identity provider and returns
// the issued impersonation token along with the consenting organization it targets
func (suite *HandlerTestSuite) supportSessionToken(t *testing.T) (string, *ent.Organization) {
	suite.registerTestHandler("POST", "v1/login", suite.h.LoginHandler)
	suite.registerTestHandler("POST", "v1/support/callback", suite.h.SupportCallbackHandler)

	oidc := newMockOIDCServer(t,
		withExpectedCode("code123"),
		withClientSecret("secret"),
		withUserInfo(suite.testUser1.UserInfo.Email, "Support Engineer", ""),
	)
	defer oidc.Close()

	suite.h.SupportAccessConfig = supportTestConfig(oidc.server.URL)
	suite.h.SupportAccessConfig.AllowedDomain = strings.SplitN(suite.testUser1.UserInfo.Email, "@", 2)[1]

	ctx := auth.WithInternalOperationContext(suite.testUser1.UserCtx)
	ctx = ent.NewContext(ctx, suite.db)
	org := suite.createConsentingOrg(ctx)

	loginBody, _ := json.Marshal(models.LoginRequest{
		Username:             "support@theopenlane.io",
		Password:             "super-secret-support-password",
		TargetOrganizationID: org.ID,
		Reason:               "assisting customer with data import",
	})
	loginReq := httptest.NewRequest(http.MethodPost, "/v1/login", strings.NewReader(string(loginBody)))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	suite.e.ServeHTTP(loginRec, loginReq)
	assert.Assert(t, is.Equal(http.StatusOK, loginRec.Code))

	cookieParts := []string{}
	var state string
	for _, c := range loginRec.Result().Cookies() {
		cookieParts = append(cookieParts, c.Name+"="+c.Value)
		if c.Name == "support_state" {
			state = c.Value
		}
		if c.Name == "support_nonce" {
			oidc.nonce = c.Value
		}
	}

	cbBody, _ := json.Marshal(models.SupportCallbackRequest{Code: "code123", State: state})
	cbReq := httptest.NewRequest(http.MethodPost, "/v1/support/callback", strings.NewReader(string(cbBody)))
	cbReq.Header.Set("Content-Type", "application/json")
	cbReq.Header.Set("Cookie", strings.Join(cookieParts, "; "))
	cbRec := httptest.NewRecorder()
	suite.e.ServeHTTP(cbRec, cbReq)
	assert.Assert(t, is.Equal(http.StatusOK, cbRec.Code))

	var cbOut models.SupportAccessResponse
	assert.NilError(t, json.NewDecoder(cbRec.Body).Decode(&cbOut))
	assert.Assert(t, len(cbOut.Token) > 0)

	return cbOut.Token, org
}

// TestSupportAccessCanQueryOrgData confirms the issued support token can read org-scoped data within the
// consented organization, without any privacy bypass in the request context
func TestSupportAccessCanQueryOrgData(t *testing.T) {
	suite := suite.setupTest(t)

	original := suite.h.SupportAccessConfig
	defer func() { suite.h.SupportAccessConfig = original }()

	token, org := suite.supportSessionToken(t)

	supportCtx := suite.supportCallerContext(token)

	setting, err := suite.db.OrganizationSetting.Query().
		Where(organizationsetting.OrganizationID(org.ID)).
		Only(supportCtx)
	assert.NilError(t, err)
	assert.Check(t, setting.AllowSupportAccess, "support should read the consenting org setting")
}

// TestSupportAccessCanUpdateOrgData confirms the issued support token can mutate org-scoped data within
// the consented organization, without any privacy bypass in the request context
func TestSupportAccessCanUpdateOrgData(t *testing.T) {
	suite := suite.setupTest(t)

	original := suite.h.SupportAccessConfig
	defer func() { suite.h.SupportAccessConfig = original }()

	token, org := suite.supportSessionToken(t)

	supportCtx := suite.supportCallerContext(token)

	// ensure org interceptor works for support
	orgMembers, err := suite.api.GetOrgMembersByOrgID(supportCtx, &testclient.OrgMembershipWhereInput{OrganizationID: &org.ID})
	assert.NilError(t, err)
	assert.Check(t, len(orgMembers.OrgMemberships.Edges) > 0)

	resp, err := suite.api.GetOrganizationSettings(supportCtx, testclient.OrganizationSettingWhereInput{})
	assert.NilError(t, err)
	assert.Check(t, len(resp.OrganizationSettings.Edges) > 0)
	assert.Check(t, is.Len(resp.OrganizationSettings.Edges, 1))

	// set for other tests
	setting := resp.OrganizationSettings.Edges[0].Node

	var updatedContact = "support-updated-contact"

	_, err = suite.api.UpdateOrganizationSetting(supportCtx, setting.ID, testclient.UpdateOrganizationSettingInput{
		BillingContact: &updatedContact,
	})
	assert.NilError(t, err)

	got, err := suite.db.OrganizationSetting.Get(supportCtx, setting.ID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(updatedContact, got.BillingContact), "support update should persist to the org setting")

	// create task to ensure assigner doesn't error on constraint
	_, err = suite.api.CreateTask(supportCtx, testclient.CreateTaskInput{
		Title: "new task from support user",
	})
	assert.NilError(t, err)
}
