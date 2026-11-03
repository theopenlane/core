package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/internal/httpserve/handlers"
	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/iam/auth"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

const (
	integrationStartPath    = "/v1/integrations/auth/start"
	integrationCallbackPath = "/v1/integrations/auth/callback"
	stateCookieName         = "state"
	orgCookieName           = "organization_id"
)

func TestStartOAuthFlow_SetsCookiesAndReturnsURL(t *testing.T) {
	suite.setupTest(t)

	suite.registerRouteOnce(http.MethodPost, integrationStartPath, suite.h.StartIntegrationAuth)

	requestCtx := auth.WithInternalOperationContext(echocontext.NewTestEchoContext().Request().Context())
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	startRec, resp := suite.startIntegrationAuth(t, user.UserCtx, handlers.IntegrationAuthStartRequest{DefinitionID: testAuthDefinitionID})

	assert.Check(t, is.Equal(http.StatusOK, startRec.Code))
	assert.Check(t, resp.Success)
	assert.Check(t, len(resp.AuthURL) > 0)
	assert.Check(t, len(resp.State) > 0)

	u, err := url.Parse(resp.AuthURL)
	assert.Check(t, err)
	assert.Check(t, len(u.Query().Get("state")) > 0)

	cookies := cookieMap(startRec.Result().Cookies())
	assert.Check(t, is.Contains(cookies, stateCookieName))
	assert.Check(t, is.Contains(cookies, orgCookieName))
}

func TestStartOAuthFlow_InvalidProvider(t *testing.T) {
	suite.setupTest(t)

	suite.registerRouteOnce(http.MethodPost, integrationStartPath, suite.h.StartIntegrationAuth)

	requestCtx := auth.WithInternalOperationContext(echocontext.NewTestEchoContext().Request().Context())
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	body, err := json.Marshal(handlers.IntegrationAuthStartRequest{DefinitionID: "def_invalid_000000000000000000", CredentialRef: testAuthCredentialRef})
	assert.Check(t, err)

	req := httptest.NewRequest(http.MethodPost, integrationStartPath, bytes.NewReader(body))
	req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(user.UserCtx))

	assert.Check(t, is.Equal(http.StatusBadRequest, rec.Code))
}

func TestStartOAuthFlow_UndeclaredOperationConfig(t *testing.T) {
	suite.setupTest(t)

	suite.registerRouteOnce(http.MethodPost, integrationStartPath, suite.h.StartIntegrationAuth)

	requestCtx := auth.WithInternalOperationContext(echocontext.NewTestEchoContext().Request().Context())
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	body, err := json.Marshal(handlers.IntegrationAuthStartRequest{
		DefinitionID:    testAuthDefinitionID,
		CredentialRef:   testAuthCredentialRef,
		OperationConfig: map[string]json.RawMessage{"undeclared_operation": json.RawMessage(`{}`)},
	})
	assert.Check(t, err)

	req := httptest.NewRequest(http.MethodPost, integrationStartPath, bytes.NewReader(body))
	req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(user.UserCtx))

	assert.Check(t, is.Equal(http.StatusBadRequest, rec.Code))
}

func TestStartOAuthFlow_Unauthorized(t *testing.T) {
	suite.setupTest(t)

	suite.registerRouteOnce(http.MethodPost, integrationStartPath, suite.h.StartIntegrationAuth)

	body, err := json.Marshal(handlers.IntegrationAuthStartRequest{DefinitionID: testAuthDefinitionID, CredentialRef: testAuthCredentialRef})
	assert.Check(t, err)

	req := httptest.NewRequest(http.MethodPost, integrationStartPath, bytes.NewReader(body))
	req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	assert.Check(t, is.Equal(http.StatusUnauthorized, rec.Code))
}

func TestHandleOAuthCallback_Success(t *testing.T) {
	suite.setupTest(t)

	suite.registerRouteOnce(http.MethodPost, integrationStartPath, suite.h.StartIntegrationAuth)
	suite.registerRouteOnce(http.MethodGet, integrationCallbackPath, suite.h.HandleIntegrationAuthCallback)

	requestCtx := auth.WithInternalOperationContext(echocontext.NewTestEchoContext().Request().Context())
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	startRec, startResp := suite.startIntegrationAuth(t, user.UserCtx, handlers.IntegrationAuthStartRequest{DefinitionID: testAuthDefinitionID})
	cookies := cookieMap(startRec.Result().Cookies())

	// OAuth state is embedded in the auth URL, not the session key (startResp.State)
	authURL, err := url.Parse(startResp.AuthURL)
	assert.Check(t, err)
	oauthState := authURL.Query().Get("state")
	assert.Check(t, len(oauthState) > 0)

	callbackReq := httptest.NewRequest(http.MethodGet, integrationCallbackPath, nil)
	query := callbackReq.URL.Query()
	query.Set("code", "test-code")
	query.Set("installation_id", "99")
	query.Set("setup_action", "install")
	query.Set("state", oauthState)
	callbackReq.URL.RawQuery = query.Encode()

	for _, name := range []string{stateCookieName, orgCookieName} {
		callbackReq.AddCookie(cookies[name])
	}
	if redirectCookie, ok := cookies["redirect_to"]; ok {
		callbackReq.AddCookie(redirectCookie)
	}

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, callbackReq.WithContext(user.UserCtx))

	assert.Check(t, is.Equal(http.StatusFound, rec.Code))

	location, err := rec.Result().Location()
	assert.NilError(t, err)
	assert.Check(t, is.Equal("http://console.example/organization-settings/integrations/"+testAuthDefinitionID, location.String()))
}

func TestHandleOAuthCallback_StateMismatch(t *testing.T) {
	suite.setupTest(t)

	suite.registerRouteOnce(http.MethodPost, integrationStartPath, suite.h.StartIntegrationAuth)
	suite.registerRouteOnce(http.MethodGet, integrationCallbackPath, suite.h.HandleIntegrationAuthCallback)

	requestCtx := auth.WithInternalOperationContext(echocontext.NewTestEchoContext().Request().Context())
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	startRec, startResp := suite.startIntegrationAuth(t, user.UserCtx, handlers.IntegrationAuthStartRequest{DefinitionID: testAuthDefinitionID})
	cookies := cookieMap(startRec.Result().Cookies())

	authURL, err := url.Parse(startResp.AuthURL)
	assert.Check(t, err)
	oauthState := authURL.Query().Get("state")
	assert.Check(t, len(oauthState) > 0)

	callbackReq := httptest.NewRequest(http.MethodGet, integrationCallbackPath, nil)
	query := callbackReq.URL.Query()
	query.Set("code", "test-code")
	query.Set("state", oauthState+"-tampered")
	callbackReq.URL.RawQuery = query.Encode()

	for _, name := range []string{stateCookieName, orgCookieName} {
		callbackReq.AddCookie(cookies[name])
	}
	if redirectCookie, ok := cookies["redirect_to"]; ok {
		callbackReq.AddCookie(redirectCookie)
	}

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, callbackReq.WithContext(user.UserCtx))

	assert.Check(t, is.Equal(http.StatusBadRequest, rec.Code))
}

func TestHandleOAuthCallback_MissingCookies(t *testing.T) {
	suite.setupTest(t)

	suite.registerRouteOnce(http.MethodGet, integrationCallbackPath, suite.h.HandleIntegrationAuthCallback)

	requestCtx := auth.WithInternalOperationContext(echocontext.NewTestEchoContext().Request().Context())
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	req := httptest.NewRequest(http.MethodGet, integrationCallbackPath+"?code=test-code&state=state", nil)
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(user.UserCtx))

	assert.Check(t, is.Equal(http.StatusBadRequest, rec.Code))
}

func TestStartOAuthFlow_SetsDefinitionBasedRedirectCookie(t *testing.T) {
	suite.setupTest(t)

	suite.registerRouteOnce(http.MethodPost, integrationStartPath, suite.h.StartIntegrationAuth)

	requestCtx := auth.WithInternalOperationContext(echocontext.NewTestEchoContext().Request().Context())
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	startRec, _ := suite.startIntegrationAuth(t, user.UserCtx, handlers.IntegrationAuthStartRequest{DefinitionID: testAuthDefinitionID})
	cookies := cookieMap(startRec.Result().Cookies())

	assert.Assert(t, is.DeepEqual("http://console.example/organization-settings/integrations/"+testAuthDefinitionID, cookies["redirect_to"].Value))
}

func TestStartOAuthFlow_SetsDefinitionBasedRedirectCookie_WithTrailingSlashConsoleURL(t *testing.T) {
	suite.setupTest(t)

	suite.h.ConsoleURL = "http://console.example/"

	suite.registerRouteOnce(http.MethodPost, integrationStartPath, suite.h.StartIntegrationAuth)

	requestCtx := auth.WithInternalOperationContext(echocontext.NewTestEchoContext().Request().Context())
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	startRec, _ := suite.startIntegrationAuth(t, user.UserCtx, handlers.IntegrationAuthStartRequest{DefinitionID: testAuthDefinitionID})
	cookies := cookieMap(startRec.Result().Cookies())

	assert.Assert(t, is.DeepEqual("http://console.example/organization-settings/integrations/"+testAuthDefinitionID, cookies["redirect_to"].Value))
}

func TestStartOAuthFlow_MissingProvider(t *testing.T) {
	suite.setupTest(t)

	suite.registerRouteOnce(http.MethodPost, integrationStartPath, suite.h.StartIntegrationAuth)

	requestCtx := auth.WithInternalOperationContext(echocontext.NewTestEchoContext().Request().Context())
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	body, err := json.Marshal(handlers.IntegrationAuthStartRequest{DefinitionID: ""})
	assert.Check(t, err)

	req := httptest.NewRequest(http.MethodPost, integrationStartPath, bytes.NewReader(body))
	req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(user.UserCtx))

	assert.Check(t, is.Equal(http.StatusBadRequest, rec.Code))
}

func TestHandleOAuthCallback_MissingState(t *testing.T) {
	suite.setupTest(t)

	suite.registerRouteOnce(http.MethodGet, integrationCallbackPath, suite.h.HandleIntegrationAuthCallback)

	requestCtx := auth.WithInternalOperationContext(echocontext.NewTestEchoContext().Request().Context())
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	req := httptest.NewRequest(http.MethodGet, integrationCallbackPath+"?code=test-code", nil)
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(user.UserCtx))

	assert.Check(t, is.Equal(http.StatusBadRequest, rec.Code))
}

func TestHandleOAuthCallback_MissingCode(t *testing.T) {
	suite.setupTest(t)

	suite.registerRouteOnce(http.MethodPost, integrationStartPath, suite.h.StartIntegrationAuth)
	suite.registerRouteOnce(http.MethodGet, integrationCallbackPath, suite.h.HandleIntegrationAuthCallback)

	requestCtx := auth.WithInternalOperationContext(echocontext.NewTestEchoContext().Request().Context())
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	startRec, startResp := suite.startIntegrationAuth(t, user.UserCtx, handlers.IntegrationAuthStartRequest{DefinitionID: testAuthDefinitionID})
	cookies := cookieMap(startRec.Result().Cookies())

	authURL, err := url.Parse(startResp.AuthURL)
	assert.Check(t, err)
	oauthState := authURL.Query().Get("state")

	req := httptest.NewRequest(http.MethodGet, integrationCallbackPath, nil)
	query := req.URL.Query()
	query.Set("state", oauthState)
	req.URL.RawQuery = query.Encode()

	for _, name := range []string{stateCookieName, orgCookieName} {
		req.AddCookie(cookies[name])
	}
	if redirectCookie, ok := cookies["redirect_to"]; ok {
		req.AddCookie(redirectCookie)
	}

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(user.UserCtx))

	assert.Check(t, is.Equal(http.StatusBadRequest, rec.Code))
}

func TestHandleOAuthCallback_MissingProviderState(t *testing.T) {
	suite.setupTest(t)

	suite.registerRouteOnce(http.MethodPost, integrationStartPath, suite.h.StartIntegrationAuth)
	suite.registerRouteOnce(http.MethodGet, integrationCallbackPath, suite.h.HandleIntegrationAuthCallback)

	requestCtx := auth.WithInternalOperationContext(echocontext.NewTestEchoContext().Request().Context())
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
	if redirectCookie, ok := cookies["redirect_to"]; ok {
		req.AddCookie(redirectCookie)
	}

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(user.UserCtx))

	assert.Check(t, is.Equal(http.StatusBadRequest, rec.Code))
}

func TestHandleOAuthCallback_InvalidCookieOrgID(t *testing.T) {
	suite.setupTest(t)

	suite.registerRouteOnce(http.MethodPost, integrationStartPath, suite.h.StartIntegrationAuth)
	suite.registerRouteOnce(http.MethodGet, integrationCallbackPath, suite.h.HandleIntegrationAuthCallback)

	requestCtx := auth.WithInternalOperationContext(echocontext.NewTestEchoContext().Request().Context())
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	startRec, startResp := suite.startIntegrationAuth(t, user.UserCtx, handlers.IntegrationAuthStartRequest{DefinitionID: testAuthDefinitionID})
	cookies := cookieMap(startRec.Result().Cookies())

	authURL, err := url.Parse(startResp.AuthURL)
	assert.Check(t, err)
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
	if redirectCookie, ok := cookies["redirect_to"]; ok {
		req.AddCookie(redirectCookie)
	}

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(user.UserCtx))

	assert.Check(t, is.Equal(http.StatusBadRequest, rec.Code))
}

func (suite *HandlerTestSuite) startIntegrationAuth(t *testing.T, ctx context.Context, request handlers.IntegrationAuthStartRequest) (*httptest.ResponseRecorder, openapi.OAuthFlowResponse) {
	t.Helper()

	if request.CredentialRef == "" {
		request.CredentialRef = testAuthCredentialRef
	}

	body, err := json.Marshal(request)
	assert.Check(t, err)

	req := httptest.NewRequest(http.MethodPost, integrationStartPath, bytes.NewReader(body))
	req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)
	req.AddCookie(&http.Cookie{Name: auth.AccessTokenCookie, Value: "access"})
	req.AddCookie(&http.Cookie{Name: auth.RefreshTokenCookie, Value: "refresh"})

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(ctx))

	assert.Check(t, is.Equal(http.StatusOK, rec.Code))

	var resp openapi.OAuthFlowResponse
	assert.Check(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	return rec, resp
}

func cookieMap(cookies []*http.Cookie) map[string]*http.Cookie {
	out := make(map[string]*http.Cookie, len(cookies))
	for _, cookie := range cookies {
		out[cookie.Name] = cookie
	}

	return out
}
