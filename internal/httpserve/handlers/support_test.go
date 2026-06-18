package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/utils/ulids"
)

// supportTestConfig returns a SupportAccessConfig pointed at the given mock OIDC server, with a known
// password and the theopenlane.io domain restriction
func supportTestConfig(issuer string) handlers.SupportAccessConfig {
	return handlers.SupportAccessConfig{
		Enabled:           true,
		Email:             "support@theopenlane.io",
		DisplayName:       "Openlane Support",
		SubjectID:         "01JSPPRT000000000000000000",
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
	setting := suite.db.OrganizationSetting.Create().SetInput(generated.CreateOrganizationSettingInput{
		AllowSupportAccess: lo.ToPtr(true),
	}).SaveX(ctx)

	org := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      ulids.New().String(),
		SettingID: &setting.ID,
	}).SaveX(ctx)

	suite.db.OrganizationSetting.UpdateOneID(setting.ID).SetOrganizationID(org.ID).ExecX(ctx)

	return org
}

func (suite *HandlerTestSuite) TestSupportAccessLoginAndCallback() {
	t := suite.T()

	loginOp := suite.createImpersonationOperation("LoginHandler", "Login handler")
	callbackOp := suite.createImpersonationOperation("SupportCallback", "Support callback handler")
	suite.registerTestHandler("POST", "v1/login", loginOp, suite.h.LoginHandler)
	suite.registerTestHandler("POST", "v1/support/callback", callbackOp, suite.h.SupportCallbackHandler)

	oidc := newMockOIDCServer(t,
		withExpectedCode("code123"),
		withClientSecret("secret"),
		withUserInfo("engineer@theopenlane.io", "Support Engineer", ""),
	)
	defer oidc.Close()

	// configure support access on the handler and restore afterwards so other tests are unaffected
	original := suite.h.SupportAccessConfig
	suite.h.SupportAccessConfig = supportTestConfig(oidc.server.URL)
	defer func() { suite.h.SupportAccessConfig = original }()

	ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)
	ctx = ent.NewContext(ctx, suite.db)

	org := suite.createConsentingOrg(ctx)

	// first factor: authenticate the support identity against the configured password
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

	require.Equal(t, http.StatusOK, loginRec.Code)

	var loginOut models.LoginReply
	require.NoError(t, json.NewDecoder(loginRec.Body).Decode(&loginOut))
	assert.True(t, loginOut.Success)
	assert.NotEmpty(t, loginOut.RedirectURI, "first factor should return the identity provider redirect")

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

	require.NotEmpty(t, state, "support_state cookie should be set")

	// second factor: complete the identity provider exchange
	cbBody, _ := json.Marshal(models.SupportCallbackRequest{Code: "code123", State: state})
	cbReq := httptest.NewRequest(http.MethodPost, "/v1/support/callback", strings.NewReader(string(cbBody)))
	cbReq.Header.Set("Content-Type", "application/json")
	cbReq.Header.Set("Cookie", strings.Join(cookieParts, "; "))
	cbRec := httptest.NewRecorder()
	suite.e.ServeHTTP(cbRec, cbReq)

	require.Equal(t, http.StatusOK, cbRec.Code)

	var cbOut models.SupportAccessReply
	require.NoError(t, json.NewDecoder(cbRec.Body).Decode(&cbOut))
	assert.True(t, cbOut.Success)
	assert.NotEmpty(t, cbOut.Token)
	assert.Equal(t, org.ID, cbOut.OrganizationID)
	assert.Equal(t, "engineer@theopenlane.io", cbOut.Impersonator)

	// the minted token must carry both identities: the virtual support user and the individual
	claims, err := suite.h.TokenManager.ValidateImpersonationToken(context.Background(), cbOut.Token)
	require.NoError(t, err)
	assert.Equal(t, "01JSPPRT000000000000000000", claims.UserID, "target is the virtual support identity")
	assert.Equal(t, "engineer@theopenlane.io", claims.ImpersonatorID, "impersonator is the individual from the IdP")
	assert.Equal(t, "support", claims.Type)
	assert.Equal(t, org.ID, claims.OrgID)
}

func (suite *HandlerTestSuite) TestSupportAccessRejectsWrongPassword() {
	t := suite.T()

	loginOp := suite.createImpersonationOperation("LoginHandler", "Login handler")
	suite.registerTestHandler("POST", "v1/login", loginOp, suite.h.LoginHandler)

	oidc := newMockOIDCServer(t)
	defer oidc.Close()

	original := suite.h.SupportAccessConfig
	suite.h.SupportAccessConfig = supportTestConfig(oidc.server.URL)
	defer func() { suite.h.SupportAccessConfig = original }()

	ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)
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

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func (suite *HandlerTestSuite) TestSupportAccessRejectsNonConsentingOrg() {
	t := suite.T()

	loginOp := suite.createImpersonationOperation("LoginHandler", "Login handler")
	suite.registerTestHandler("POST", "v1/login", loginOp, suite.h.LoginHandler)

	oidc := newMockOIDCServer(t)
	defer oidc.Close()

	original := suite.h.SupportAccessConfig
	suite.h.SupportAccessConfig = supportTestConfig(oidc.server.URL)
	defer func() { suite.h.SupportAccessConfig = original }()

	ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)
	ctx = ent.NewContext(ctx, suite.db)

	// org that has NOT consented to support access
	setting := suite.db.OrganizationSetting.Create().SetInput(generated.CreateOrganizationSettingInput{}).SaveX(ctx)
	org := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      ulids.New().String(),
		SettingID: &setting.ID,
	}).SaveX(ctx)
	suite.db.OrganizationSetting.UpdateOneID(setting.ID).SetOrganizationID(org.ID).ExecX(ctx)

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

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// TestSupportAccessRejectsDomainMismatch ensures the second factor rejects an individual outside the
// configured domain even when the first factor and consent succeed
func (suite *HandlerTestSuite) TestSupportAccessRejectsDomainMismatch() {
	t := suite.T()

	loginOp := suite.createImpersonationOperation("LoginHandler", "Login handler")
	callbackOp := suite.createImpersonationOperation("SupportCallback", "Support callback handler")
	suite.registerTestHandler("POST", "v1/login", loginOp, suite.h.LoginHandler)
	suite.registerTestHandler("POST", "v1/support/callback", callbackOp, suite.h.SupportCallbackHandler)

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

	ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)
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
	require.Equal(t, http.StatusOK, loginRec.Code)

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

	assert.Equal(t, http.StatusForbidden, cbRec.Code)
}
