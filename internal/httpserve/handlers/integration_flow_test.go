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
	v2definition "github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

const (
	integrationStartPath    = "/v1/integrations/oauth/start"
	integrationCallbackPath = "/v1/integrations/oauth/callback"
	stateCookieName         = "oauth_state"
	orgCookieName           = "oauth_org_id"
	userCookieName          = "oauth_user_id"
)

func (suite *HandlerTestSuite) TestStartOAuthFlow_SetsCookiesAndReturnsURL() {
	t := suite.T()

	op := suite.createImpersonationOperation("StartIntegrationOAuth", "Start integration OAuth flow")
	suite.registerRouteOnce(http.MethodPost, integrationStartPath, op, suite.h.StartOAuthFlow)

	requestCtx := privacy.DecisionContext(echocontext.NewTestEchoContext().Request().Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	startRec, resp := suite.startOAuthFlow(t, user.UserCtx, handlers.OAuthV2FlowRequest{DefinitionID: testOAuthDefinitionID})

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
	require.Contains(t, cookies, userCookieName)

	assert.Equal(t, http.SameSiteLaxMode, cookies[stateCookieName].SameSite)
	assert.Equal(t, http.SameSiteLaxMode, cookies[orgCookieName].SameSite)
	assert.Equal(t, http.SameSiteLaxMode, cookies[userCookieName].SameSite)
}

func (suite *HandlerTestSuite) TestStartOAuthFlow_InvalidProvider() {
	t := suite.T()

	op := suite.createImpersonationOperation("StartIntegrationOAuthInvalid", "Start integration OAuth flow")
	suite.registerRouteOnce(http.MethodPost, integrationStartPath, op, suite.h.StartOAuthFlow)

	requestCtx := privacy.DecisionContext(echocontext.NewTestEchoContext().Request().Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	body, err := json.Marshal(handlers.OAuthV2FlowRequest{DefinitionID: "def_invalid_000000000000000000"})
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
	suite.registerRouteOnce(http.MethodPost, integrationStartPath, op, suite.h.StartOAuthFlow)

	body, err := json.Marshal(handlers.OAuthV2FlowRequest{DefinitionID: testOAuthDefinitionID})
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
	suite.registerRouteOnce(http.MethodPost, integrationStartPath, startOp, suite.h.StartOAuthFlow)

	callbackOp := suite.createImpersonationOperation("HandleIntegrationOAuthCallback", "Handle integration OAuth callback")
	suite.registerRouteOnce(http.MethodGet, integrationCallbackPath, callbackOp, suite.h.HandleOAuthCallback)

	requestCtx := privacy.DecisionContext(echocontext.NewTestEchoContext().Request().Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	startRec, startResp := suite.startOAuthFlow(t, user.UserCtx, handlers.OAuthV2FlowRequest{DefinitionID: testOAuthDefinitionID})
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

	for _, name := range []string{stateCookieName, orgCookieName, userCookieName} {
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
	suite.registerRouteOnce(http.MethodPost, integrationStartPath, startOp, suite.h.StartOAuthFlow)

	callbackOp := suite.createImpersonationOperation("HandleIntegrationOAuthStateMismatch", "Handle integration OAuth callback")
	suite.registerRouteOnce(http.MethodGet, integrationCallbackPath, callbackOp, suite.h.HandleOAuthCallback)

	requestCtx := privacy.DecisionContext(echocontext.NewTestEchoContext().Request().Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	startRec, startResp := suite.startOAuthFlow(t, user.UserCtx, handlers.OAuthV2FlowRequest{DefinitionID: testOAuthDefinitionID})
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

	for _, name := range []string{stateCookieName, orgCookieName, userCookieName} {
		callbackReq.AddCookie(cookies[name])
	}

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, callbackReq.WithContext(user.UserCtx))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func (suite *HandlerTestSuite) TestHandleOAuthCallback_MissingCookies() {
	t := suite.T()

	callbackOp := suite.createImpersonationOperation("HandleIntegrationOAuthMissingCookies", "Handle integration OAuth callback")
	suite.registerRouteOnce(http.MethodGet, integrationCallbackPath, callbackOp, suite.h.HandleOAuthCallback)

	requestCtx := privacy.DecisionContext(echocontext.NewTestEchoContext().Request().Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	req := httptest.NewRequest(http.MethodGet, integrationCallbackPath+"?code=test-code&state=state", nil)
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(user.UserCtx))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestHandleOAuthCallback_RedirectsWhenConfigured() {
	t := suite.T()

	startOp := suite.createImpersonationOperation("StartIntegrationOAuthRedirect", "Start integration OAuth flow")
	suite.registerRouteOnce(http.MethodPost, integrationStartPath, startOp, suite.h.StartOAuthFlow)

	callbackOp := suite.createImpersonationOperation("HandleIntegrationOAuthRedirect", "Handle integration OAuth callback")
	suite.registerRouteOnce(http.MethodGet, integrationCallbackPath, callbackOp, suite.h.HandleOAuthCallback)

	restore := suite.withDefinitionRuntime(t, []v2definition.Builder{v2definition.Builder(buildTestOAuthDefinition)}, "https://console.openlane.io/integrations")
	defer restore()

	requestCtx := privacy.DecisionContext(echocontext.NewTestEchoContext().Request().Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	startRec, startResp := suite.startOAuthFlow(t, user.UserCtx, handlers.OAuthV2FlowRequest{DefinitionID: testOAuthDefinitionID})
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

	for _, name := range []string{stateCookieName, orgCookieName, userCookieName} {
		req.AddCookie(cookies[name])
	}

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(user.UserCtx))

	assert.Equal(t, http.StatusFound, rec.Code)
	location := rec.Header().Get("Location")
	assert.Contains(t, location, "provider=test-oauth")
	assert.Contains(t, location, "status=success")
}

func (suite *HandlerTestSuite) TestStartOAuthFlow_MissingProvider() {
	t := suite.T()

	op := suite.createImpersonationOperation("StartIntegrationOAuthMissingProvider", "Start integration OAuth flow")
	suite.registerRouteOnce(http.MethodPost, integrationStartPath, op, suite.h.StartOAuthFlow)

	requestCtx := privacy.DecisionContext(echocontext.NewTestEchoContext().Request().Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	body, err := json.Marshal(handlers.OAuthV2FlowRequest{DefinitionID: ""})
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
	suite.registerRouteOnce(http.MethodGet, integrationCallbackPath, callbackOp, suite.h.HandleOAuthCallback)

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
	suite.registerRouteOnce(http.MethodPost, integrationStartPath, startOp, suite.h.StartOAuthFlow)

	callbackOp := suite.createImpersonationOperation("HandleIntegrationOAuthMissingCode", "Handle integration OAuth callback")
	suite.registerRouteOnce(http.MethodGet, integrationCallbackPath, callbackOp, suite.h.HandleOAuthCallback)

	requestCtx := privacy.DecisionContext(echocontext.NewTestEchoContext().Request().Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	startRec, startResp := suite.startOAuthFlow(t, user.UserCtx, handlers.OAuthV2FlowRequest{DefinitionID: testOAuthDefinitionID})
	cookies := cookieMap(startRec.Result().Cookies())

	authURL, err := url.Parse(startResp.AuthURL)
	require.NoError(t, err)
	oauthState := authURL.Query().Get("state")

	req := httptest.NewRequest(http.MethodGet, integrationCallbackPath, nil)
	query := req.URL.Query()
	query.Set("state", oauthState)
	req.URL.RawQuery = query.Encode()

	for _, name := range []string{stateCookieName, orgCookieName, userCookieName} {
		req.AddCookie(cookies[name])
	}

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(user.UserCtx))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestHandleOAuthCallback_InvalidCookieOrgID() {
	t := suite.T()

	startOp := suite.createImpersonationOperation("StartIntegrationOAuthInvalidOrg", "Start integration OAuth flow")
	suite.registerRouteOnce(http.MethodPost, integrationStartPath, startOp, suite.h.StartOAuthFlow)

	callbackOp := suite.createImpersonationOperation("HandleIntegrationOAuthInvalidOrg", "Handle integration OAuth callback")
	suite.registerRouteOnce(http.MethodGet, integrationCallbackPath, callbackOp, suite.h.HandleOAuthCallback)

	requestCtx := privacy.DecisionContext(echocontext.NewTestEchoContext().Request().Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	startRec, startResp := suite.startOAuthFlow(t, user.UserCtx, handlers.OAuthV2FlowRequest{DefinitionID: testOAuthDefinitionID})
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

	for _, name := range []string{stateCookieName, orgCookieName, userCookieName} {
		req.AddCookie(cookies[name])
	}

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(user.UserCtx))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) startOAuthFlow(t *testing.T, ctx context.Context, request handlers.OAuthV2FlowRequest) (*httptest.ResponseRecorder, openapi.OAuthFlowResponse) {
	t.Helper()

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

func cookieMap(cookies []*http.Cookie) map[string]*http.Cookie {
	out := make(map[string]*http.Cookie, len(cookies))
	for _, cookie := range cookies {
		out[cookie.Name] = cookie
	}

	return out
}
