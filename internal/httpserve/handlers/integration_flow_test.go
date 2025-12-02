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

	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/ent/generated/privacy"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/iam/auth"
	models "github.com/theopenlane/shared/openapi"
	"github.com/theopenlane/utils/rout"
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

	rec, resp := suite.startOAuthFlow(t, user.UserCtx, models.OAuthFlowRequest{Provider: "github"})

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, resp.Success)
	assert.NotEmpty(t, resp.AuthURL)
	assert.NotEmpty(t, resp.State)

	u, err := url.Parse(resp.AuthURL)
	require.NoError(t, err)
	assert.Equal(t, resp.State, u.Query().Get("state"))

	cookies := cookieMap(rec.Result().Cookies())
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

	body, err := json.Marshal(models.OAuthFlowRequest{Provider: "invalid"})
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

	body, err := json.Marshal(models.OAuthFlowRequest{Provider: "github"})
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

	startRec, resp := suite.startOAuthFlow(t, user.UserCtx, models.OAuthFlowRequest{Provider: "github"})
	cookies := cookieMap(startRec.Result().Cookies())

	callbackReq := httptest.NewRequest(http.MethodGet, integrationCallbackPath, nil)
	query := callbackReq.URL.Query()
	query.Set("code", "test-code")
	query.Set("state", resp.State)
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

	startRec, resp := suite.startOAuthFlow(t, user.UserCtx, models.OAuthFlowRequest{Provider: "github"})
	cookies := cookieMap(startRec.Result().Cookies())

	callbackReq := httptest.NewRequest(http.MethodGet, integrationCallbackPath, nil)
	query := callbackReq.URL.Query()
	query.Set("code", "test-code")
	query.Set("state", resp.State+"-tampered")
	callbackReq.URL.RawQuery = query.Encode()

	for _, name := range []string{stateCookieName, orgCookieName, userCookieName} {
		callbackReq.AddCookie(cookies[name])
	}

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, callbackReq.WithContext(user.UserCtx))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
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

	originalRedirect := suite.h.IntegrationOauthProvider.SuccessRedirectURL
	suite.h.IntegrationOauthProvider.SuccessRedirectURL = "https://console.openlane.io/integrations"
	defer func() {
		suite.h.IntegrationOauthProvider.SuccessRedirectURL = originalRedirect
	}()

	requestCtx := privacy.DecisionContext(echocontext.NewTestEchoContext().Request().Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	startRec, resp := suite.startOAuthFlow(t, user.UserCtx, models.OAuthFlowRequest{Provider: "github"})
	cookies := cookieMap(startRec.Result().Cookies())

	req := httptest.NewRequest(http.MethodGet, integrationCallbackPath, nil)
	query := req.URL.Query()
	query.Set("code", "test-code")
	query.Set("state", resp.State)
	req.URL.RawQuery = query.Encode()

	for _, name := range []string{stateCookieName, orgCookieName, userCookieName} {
		req.AddCookie(cookies[name])
	}

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(user.UserCtx))

	assert.Equal(t, http.StatusFound, rec.Code)
	location := rec.Header().Get("Location")
	assert.Contains(t, location, "provider=github")
	assert.Contains(t, location, "status=success")
}

func (suite *HandlerTestSuite) startOAuthFlow(t *testing.T, ctx context.Context, request models.OAuthFlowRequest) (*httptest.ResponseRecorder, models.OAuthFlowResponse) {
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

	var resp models.OAuthFlowResponse
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

func (suite *HandlerTestSuite) TestStartOAuthFlow_MissingProvider() {
	t := suite.T()

	op := suite.createImpersonationOperation("StartIntegrationOAuthMissingProvider", "Start integration OAuth flow")
	suite.registerRouteOnce(http.MethodPost, integrationStartPath, op, suite.h.StartOAuthFlow)

	requestCtx := privacy.DecisionContext(echocontext.NewTestEchoContext().Request().Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	body, err := json.Marshal(models.OAuthFlowRequest{Provider: ""})
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

	startRec, resp := suite.startOAuthFlow(t, user.UserCtx, models.OAuthFlowRequest{Provider: "github"})
	cookies := cookieMap(startRec.Result().Cookies())

	req := httptest.NewRequest(http.MethodGet, integrationCallbackPath, nil)
	query := req.URL.Query()
	query.Set("state", resp.State)
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

	startRec, resp := suite.startOAuthFlow(t, user.UserCtx, models.OAuthFlowRequest{Provider: "github"})
	cookies := cookieMap(startRec.Result().Cookies())

	cookies[orgCookieName].Value = "invalid-org-id"

	req := httptest.NewRequest(http.MethodGet, integrationCallbackPath, nil)
	query := req.URL.Query()
	query.Set("code", "test-code")
	query.Set("state", resp.State)
	req.URL.RawQuery = query.Encode()

	for _, name := range []string{stateCookieName, orgCookieName, userCookieName} {
		req.AddCookie(cookies[name])
	}

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(user.UserCtx))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
