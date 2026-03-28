package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

const (
	integrationStartPath    = "/v1/integrations/auth/start"
	integrationCallbackPath = "/v1/integrations/auth/callback"
	stateCookieName         = "state"
	orgCookieName           = "organization_id"
)

func (suite *HandlerTestSuite) TestStartOAuthFlow_SetsCookiesAndReturnsURL() {
	t := suite.T()

	op := suite.createImpersonationOperation("StartIntegrationOAuth", "Start integration OAuth flow")
	suite.registerRouteOnce(http.MethodPost, integrationStartPath, op, suite.h.StartIntegrationAuth)

	requestCtx := privacy.DecisionContext(echocontext.NewTestEchoContext().Request().Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	startRec, resp := suite.startIntegrationAuth(t, user.UserCtx, handlers.IntegrationAuthStartRequest{DefinitionID: testAuthDefinitionID})

	assert.Equal(t, http.StatusOK, startRec.Code)
	assert.True(t, resp.Success)
	assert.NotEmpty(t, resp.AuthURL)
	assert.NotEmpty(t, resp.State)

	u, err := url.Parse(resp.AuthURL)
	require.NoError(t, err)
	assert.NotEmpty(t, u.Query().Get("state"))

	cookies := cookieMap(startRec.Result().Cookies())
	require.Contains(t, cookies, stateCookieName)
	require.Contains(t, cookies, orgCookieName)
}

func (suite *HandlerTestSuite) TestStartOAuthFlow_InvalidProvider() {
	t := suite.T()

	op := suite.createImpersonationOperation("StartIntegrationOAuthInvalid", "Start integration OAuth flow")
	suite.registerRouteOnce(http.MethodPost, integrationStartPath, op, suite.h.StartIntegrationAuth)

	requestCtx := privacy.DecisionContext(echocontext.NewTestEchoContext().Request().Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	body, err := json.Marshal(handlers.IntegrationAuthStartRequest{DefinitionID: "def_invalid_000000000000000000", CredentialRef: testAuthCredentialRef})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, integrationStartPath, bytes.NewReader(body))
	req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(user.UserCtx))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestStartOAuthFlow_Unauthorized() {
	t := suite.T()

	op := suite.createImpersonationOperation("StartIntegrationOAuthUnauthorized", "Start integration OAuth flow")
	suite.registerRouteOnce(http.MethodPost, integrationStartPath, op, suite.h.StartIntegrationAuth)

	body, err := json.Marshal(handlers.IntegrationAuthStartRequest{DefinitionID: testAuthDefinitionID, CredentialRef: testAuthCredentialRef})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, integrationStartPath, bytes.NewReader(body))
	req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func (suite *HandlerTestSuite) TestHandleOAuthCallback_Success() {
	t := suite.T()

	startOp := suite.createImpersonationOperation("StartIntegrationOAuthCallback", "Start integration OAuth flow")
	suite.registerRouteOnce(http.MethodPost, integrationStartPath, startOp, suite.h.StartIntegrationAuth)

	callbackOp := suite.createImpersonationOperation("HandleIntegrationOAuthCallback", "Handle integration OAuth callback")
	suite.registerRouteOnce(http.MethodGet, integrationCallbackPath, callbackOp, suite.h.HandleIntegrationAuthCallback)

	requestCtx := privacy.DecisionContext(echocontext.NewTestEchoContext().Request().Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	startRec, startResp := suite.startIntegrationAuth(t, user.UserCtx, handlers.IntegrationAuthStartRequest{DefinitionID: testAuthDefinitionID})
	cookies := cookieMap(startRec.Result().Cookies())

	// OAuth state is embedded in the auth URL, not the session key (startResp.State)
	authURL, err := url.Parse(startResp.AuthURL)
	require.NoError(t, err)
	oauthState := authURL.Query().Get("state")
	require.NotEmpty(t, oauthState)

	callbackReq := httptest.NewRequest(http.MethodGet, integrationCallbackPath, nil)
	query := callbackReq.URL.Query()
	query.Set("code", "test-code")
	query.Set("state", oauthState)
	callbackReq.URL.RawQuery = query.Encode()

	for _, name := range []string{stateCookieName, orgCookieName} {
		callbackReq.AddCookie(cookies[name])
	}

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, callbackReq.WithContext(user.UserCtx))

	assert.Equal(t, http.StatusOK, rec.Code)

	var out rout.Reply
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.True(t, out.Success)
}

func (suite *HandlerTestSuite) TestHandleOAuthCallback_StateMismatch() {
	t := suite.T()

	startOp := suite.createImpersonationOperation("StartIntegrationOAuthStateMismatch", "Start integration OAuth flow")
	suite.registerRouteOnce(http.MethodPost, integrationStartPath, startOp, suite.h.StartIntegrationAuth)

	callbackOp := suite.createImpersonationOperation("HandleIntegrationOAuthStateMismatch", "Handle integration OAuth callback")
	suite.registerRouteOnce(http.MethodGet, integrationCallbackPath, callbackOp, suite.h.HandleIntegrationAuthCallback)

	requestCtx := privacy.DecisionContext(echocontext.NewTestEchoContext().Request().Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	startRec, startResp := suite.startIntegrationAuth(t, user.UserCtx, handlers.IntegrationAuthStartRequest{DefinitionID: testAuthDefinitionID})
	cookies := cookieMap(startRec.Result().Cookies())

	authURL, err := url.Parse(startResp.AuthURL)
	require.NoError(t, err)
	oauthState := authURL.Query().Get("state")
	require.NotEmpty(t, oauthState)

	callbackReq := httptest.NewRequest(http.MethodGet, integrationCallbackPath, nil)
	query := callbackReq.URL.Query()
	query.Set("code", "test-code")
	query.Set("state", oauthState+"-tampered")
	callbackReq.URL.RawQuery = query.Encode()

	for _, name := range []string{stateCookieName, orgCookieName} {
		callbackReq.AddCookie(cookies[name])
	}

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, callbackReq.WithContext(user.UserCtx))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestHandleOAuthCallback_MissingCookies() {
	t := suite.T()

	callbackOp := suite.createImpersonationOperation("HandleIntegrationOAuthMissingCookies", "Handle integration OAuth callback")
	suite.registerRouteOnce(http.MethodGet, integrationCallbackPath, callbackOp, suite.h.HandleIntegrationAuthCallback)

	requestCtx := privacy.DecisionContext(echocontext.NewTestEchoContext().Request().Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	req := httptest.NewRequest(http.MethodGet, integrationCallbackPath+"?code=test-code&state=state", nil)
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(user.UserCtx))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestHandleOAuthCallback_ReturnsSuccessResponse() {
	t := suite.T()

	startOp := suite.createImpersonationOperation("StartIntegrationOAuthRedirect", "Start integration OAuth flow")
	suite.registerRouteOnce(http.MethodPost, integrationStartPath, startOp, suite.h.StartIntegrationAuth)

	callbackOp := suite.createImpersonationOperation("HandleIntegrationOAuthRedirect", "Handle integration OAuth callback")
	suite.registerRouteOnce(http.MethodGet, integrationCallbackPath, callbackOp, suite.h.HandleIntegrationAuthCallback)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{registry.Builder(buildTestOAuthDefinition)})
	defer restore()

	requestCtx := privacy.DecisionContext(echocontext.NewTestEchoContext().Request().Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	startRec, startResp := suite.startIntegrationAuth(t, user.UserCtx, handlers.IntegrationAuthStartRequest{DefinitionID: testAuthDefinitionID})
	cookies := cookieMap(startRec.Result().Cookies())

	authURL, err := url.Parse(startResp.AuthURL)
	require.NoError(t, err)
	oauthState := authURL.Query().Get("state")
	require.NotEmpty(t, oauthState)

	req := httptest.NewRequest(http.MethodGet, integrationCallbackPath, nil)
	query := req.URL.Query()
	query.Set("code", "test-code")
	query.Set("state", oauthState)
	req.URL.RawQuery = query.Encode()

	for _, name := range []string{stateCookieName, orgCookieName} {
		req.AddCookie(cookies[name])
	}

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(user.UserCtx))

	assert.Equal(t, http.StatusOK, rec.Code)

	var out rout.Reply
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.True(t, out.Success)
}

func (suite *HandlerTestSuite) TestStartOAuthFlow_MissingProvider() {
	t := suite.T()

	op := suite.createImpersonationOperation("StartIntegrationOAuthMissingProvider", "Start integration OAuth flow")
	suite.registerRouteOnce(http.MethodPost, integrationStartPath, op, suite.h.StartIntegrationAuth)

	requestCtx := privacy.DecisionContext(echocontext.NewTestEchoContext().Request().Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	body, err := json.Marshal(handlers.IntegrationAuthStartRequest{DefinitionID: ""})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, integrationStartPath, bytes.NewReader(body))
	req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(user.UserCtx))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestHandleOAuthCallback_MissingState() {
	t := suite.T()

	callbackOp := suite.createImpersonationOperation("HandleIntegrationOAuthMissingState", "Handle integration OAuth callback")
	suite.registerRouteOnce(http.MethodGet, integrationCallbackPath, callbackOp, suite.h.HandleIntegrationAuthCallback)

	requestCtx := privacy.DecisionContext(echocontext.NewTestEchoContext().Request().Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	req := httptest.NewRequest(http.MethodGet, integrationCallbackPath+"?code=test-code", nil)
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(user.UserCtx))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestHandleOAuthCallback_MissingCode() {
	t := suite.T()

	startOp := suite.createImpersonationOperation("StartIntegrationOAuthMissingCode", "Start integration OAuth flow")
	suite.registerRouteOnce(http.MethodPost, integrationStartPath, startOp, suite.h.StartIntegrationAuth)

	callbackOp := suite.createImpersonationOperation("HandleIntegrationOAuthMissingCode", "Handle integration OAuth callback")
	suite.registerRouteOnce(http.MethodGet, integrationCallbackPath, callbackOp, suite.h.HandleIntegrationAuthCallback)

	requestCtx := privacy.DecisionContext(echocontext.NewTestEchoContext().Request().Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	startRec, startResp := suite.startIntegrationAuth(t, user.UserCtx, handlers.IntegrationAuthStartRequest{DefinitionID: testAuthDefinitionID})
	cookies := cookieMap(startRec.Result().Cookies())

	authURL, err := url.Parse(startResp.AuthURL)
	require.NoError(t, err)
	oauthState := authURL.Query().Get("state")

	req := httptest.NewRequest(http.MethodGet, integrationCallbackPath, nil)
	query := req.URL.Query()
	query.Set("state", oauthState)
	req.URL.RawQuery = query.Encode()

	for _, name := range []string{stateCookieName, orgCookieName} {
		req.AddCookie(cookies[name])
	}

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(user.UserCtx))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestHandleOAuthCallback_MissingProviderState() {
	t := suite.T()

	startOp := suite.createImpersonationOperation("StartIntegrationOAuthMissingProviderState", "Start integration OAuth flow")
	suite.registerRouteOnce(http.MethodPost, integrationStartPath, startOp, suite.h.StartIntegrationAuth)

	callbackOp := suite.createImpersonationOperation("HandleIntegrationOAuthMissingProviderState", "Handle integration OAuth callback")
	suite.registerRouteOnce(http.MethodGet, integrationCallbackPath, callbackOp, suite.h.HandleIntegrationAuthCallback)

	requestCtx := privacy.DecisionContext(echocontext.NewTestEchoContext().Request().Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	startRec, _ := suite.startIntegrationAuth(t, user.UserCtx, handlers.IntegrationAuthStartRequest{DefinitionID: testAuthDefinitionID})
	cookies := cookieMap(startRec.Result().Cookies())

	req := httptest.NewRequest(http.MethodGet, integrationCallbackPath, nil)
	query := req.URL.Query()
	query.Set("code", "test-code")
	req.URL.RawQuery = query.Encode()

	for _, name := range []string{stateCookieName, orgCookieName} {
		req.AddCookie(cookies[name])
	}

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(user.UserCtx))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestHandleOAuthCallback_InvalidCookieOrgID() {
	t := suite.T()

	startOp := suite.createImpersonationOperation("StartIntegrationOAuthInvalidOrg", "Start integration OAuth flow")
	suite.registerRouteOnce(http.MethodPost, integrationStartPath, startOp, suite.h.StartIntegrationAuth)

	callbackOp := suite.createImpersonationOperation("HandleIntegrationOAuthInvalidOrg", "Handle integration OAuth callback")
	suite.registerRouteOnce(http.MethodGet, integrationCallbackPath, callbackOp, suite.h.HandleIntegrationAuthCallback)

	requestCtx := privacy.DecisionContext(echocontext.NewTestEchoContext().Request().Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	startRec, startResp := suite.startIntegrationAuth(t, user.UserCtx, handlers.IntegrationAuthStartRequest{DefinitionID: testAuthDefinitionID})
	cookies := cookieMap(startRec.Result().Cookies())

	authURL, err := url.Parse(startResp.AuthURL)
	require.NoError(t, err)
	oauthState := authURL.Query().Get("state")

	cookies[orgCookieName].Value = "invalid-org-id"

	req := httptest.NewRequest(http.MethodGet, integrationCallbackPath, nil)
	query := req.URL.Query()
	query.Set("code", "test-code")
	query.Set("state", oauthState)
	req.URL.RawQuery = query.Encode()

	for _, name := range []string{stateCookieName, orgCookieName} {
		req.AddCookie(cookies[name])
	}

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(user.UserCtx))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) startIntegrationAuth(t *testing.T, ctx context.Context, request handlers.IntegrationAuthStartRequest) (*httptest.ResponseRecorder, openapi.OAuthFlowResponse) {
	t.Helper()

	if request.CredentialRef == (types.CredentialSlotID{}) {
		request.CredentialRef = testAuthCredentialRef
	}

	body, err := json.Marshal(request)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, integrationStartPath, bytes.NewReader(body))
	req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)
	req.AddCookie(&http.Cookie{Name: auth.AccessTokenCookie, Value: "access"})
	req.AddCookie(&http.Cookie{Name: auth.RefreshTokenCookie, Value: "refresh"})

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(ctx))

	require.Equal(t, http.StatusOK, rec.Code)

	var resp openapi.OAuthFlowResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	return rec, resp
}

func (suite *HandlerTestSuite) completeOAuthInstallation(t *testing.T, ctx context.Context) string {
	t.Helper()

	startRec, startResp := suite.startIntegrationAuth(t, ctx, handlers.IntegrationAuthStartRequest{DefinitionID: testAuthDefinitionID})
	cookies := cookieMap(startRec.Result().Cookies())

	authURL, err := url.Parse(startResp.AuthURL)
	require.NoError(t, err)
	oauthState := authURL.Query().Get("state")
	require.NotEmpty(t, oauthState)

	req := httptest.NewRequest(http.MethodGet, integrationCallbackPath, nil)
	query := req.URL.Query()
	query.Set("code", "test-code")
	query.Set("state", oauthState)
	req.URL.RawQuery = query.Encode()

	for _, name := range []string{stateCookieName, orgCookieName} {
		req.AddCookie(cookies[name])
	}

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(ctx))
	require.Equal(t, http.StatusOK, rec.Code)

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	require.NoError(t, err)

	record := suite.db.Integration.Query().
		Where(
			integration.OwnerIDEQ(orgID),
			integration.DefinitionIDEQ(testAuthDefinitionID),
		).
		OnlyX(ctx)

	return record.ID
}

func cookieMap(cookies []*http.Cookie) map[string]*http.Cookie {
	out := make(map[string]*http.Cookie, len(cookies))
	for _, cookie := range cookies {
		out[cookie.Name] = cookie
	}

	return out
}
