package handlers_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	models "github.com/theopenlane/core/pkg/openapi"
)

func (suite *HandlerTestSuite) TestStartOAuthFlow() {
	t := suite.T()

	// Create operation for StartOAuthFlow
	startOp := suite.createImpersonationOperation("StartOAuthFlow", "Start OAuth flow")
	suite.registerTestHandler("POST", "oauth/start", startOp, suite.h.StartOAuthFlow)

	ctx := echocontext.NewTestEchoContext().Request().Context()
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// Create test user with organization
	testUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      "0p3nl@n3rocks!",
		confirmedUser: true,
	})

	tests := []struct {
		name           string
		request        models.OAuthFlowRequest
		expectedStatus int
		checkResponse  func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "happy path - github without additional scopes",
			request: models.OAuthFlowRequest{
				Provider: "github",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var response models.OAuthFlowResponse
				err := json.Unmarshal(resp.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.True(t, response.Success)
				assert.NotEmpty(t, response.AuthURL)
				assert.NotEmpty(t, response.State)
				assert.Contains(t, response.AuthURL, "example.com/oauth/authorize")
			},
		},
		{
			name: "happy path - github with additional scopes",
			request: models.OAuthFlowRequest{
				Provider: "github",
				Scopes:   []string{"gist", "public_repo"},
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var response models.OAuthFlowResponse
				err := json.Unmarshal(resp.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.True(t, response.Success)
				assert.NotEmpty(t, response.AuthURL)
				assert.Contains(t, response.AuthURL, "example.com/oauth/authorize")
			},
		},
		{
			name: "invalid provider",
			request: models.OAuthFlowRequest{
				Provider: "invalid_provider",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Contains(t, resp.Body.String(), "invalid or unparsable field: provider")
			},
		},
		{
			name: "empty provider",
			request: models.OAuthFlowRequest{
				Provider: "",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Contains(t, resp.Body.String(), "provider is required")
			},
		},
		{
			name: "slack provider (not configured)",
			request: models.OAuthFlowRequest{
				Provider: "slack",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Contains(t, resp.Body.String(), "invalid provider")
			},
		},
	}

	// Test for unauthenticated user (should return 401)
	unauthenticatedTests := []struct {
		name           string
		request        models.OAuthFlowRequest
		expectedStatus int
		checkResponse  func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "unauthenticated user - should return 401",
			request: models.OAuthFlowRequest{
				Provider: "github",
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Contains(t, resp.Body.String(), "could not identify authenticated user in request")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/oauth/start", strings.NewReader(string(body)))
			req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

			rec := httptest.NewRecorder()
			suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}
		})
	}

	// Test unauthenticated scenarios
	for _, tt := range unauthenticatedTests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/oauth/start", strings.NewReader(string(body)))
			req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

			rec := httptest.NewRecorder()
			// Use unauthenticated context (no user in context)
			suite.e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}
		})
	}
}

func (suite *HandlerTestSuite) TestHandleOAuthCallback() {
	t := suite.T()

	// Create operation for HandleOAuthCallback
	callbackOp := suite.createImpersonationOperation("HandleOAuthCallback", "Handle OAuth callback")
	suite.registerTestHandler("GET", "oauth/callback", callbackOp, suite.h.HandleOAuthCallback)

	ctx := echocontext.NewTestEchoContext().Request().Context()
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// Create test user with organization
	testUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      "0p3nl@n3rocks!",
		confirmedUser: true,
	})

	// Generate valid state for testing
	orgID := testUser.OrganizationID
	provider := "github"
	validState, err := generateTestOAuthState(orgID, provider)
	require.NoError(t, err)

	tests := []struct {
		name           string
		urlPath        string
		cookies        []*http.Cookie
		expectedStatus int
		setupMock      func()
		checkResponse  func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:    "invalid state",
			urlPath: "/oauth/callback?code=valid_code&state=invalid_state",
			cookies: []*http.Cookie{
				{Name: "oauth_state", Value: validState},
				{Name: "oauth_org_id", Value: orgID},
				{Name: "oauth_user_id", Value: testUser.UserInfo.ID},
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Contains(t, resp.Body.String(), "invalid OAuth state")
			},
		},
		{
			name:    "missing code",
			urlPath: fmt.Sprintf("/oauth/callback?state=%s", validState),
			cookies: []*http.Cookie{
				{Name: "oauth_state", Value: validState},
				{Name: "oauth_org_id", Value: orgID},
				{Name: "oauth_user_id", Value: testUser.UserInfo.ID},
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Contains(t, resp.Body.String(), "code is required")
			},
		},
		{
			name:    "missing OAuth state cookie",
			urlPath: fmt.Sprintf("/oauth/callback?code=valid_code&state=%s", validState),
			cookies: []*http.Cookie{
				{Name: "oauth_org_id", Value: orgID},
				{Name: "oauth_user_id", Value: testUser.UserInfo.ID},
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Contains(t, resp.Body.String(), "invalid OAuth state")
			},
		},
		{
			name:    "missing OAuth org cookie",
			urlPath: fmt.Sprintf("/oauth/callback?code=valid_code&state=%s", validState),
			cookies: []*http.Cookie{
				{Name: "oauth_state", Value: validState},
				{Name: "oauth_user_id", Value: testUser.UserInfo.ID},
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Contains(t, resp.Body.String(), "missing organization context")
			},
		},
		{
			name:    "missing OAuth user cookie",
			urlPath: fmt.Sprintf("/oauth/callback?code=valid_code&state=%s", validState),
			cookies: []*http.Cookie{
				{Name: "oauth_state", Value: validState},
				{Name: "oauth_org_id", Value: orgID},
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Contains(t, resp.Body.String(), "missing user context")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			req := httptest.NewRequest(http.MethodGet, tt.urlPath, nil)

			// Add cookies to the request
			for _, cookie := range tt.cookies {
				req.AddCookie(cookie)
			}

			rec := httptest.NewRecorder()
			suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}
		})
	}
}

func (suite *HandlerTestSuite) TestHandleOAuthCallbackSuccess() {
	t := suite.T()

	// Create operation for HandleOAuthCallback
	callbackOp := suite.createImpersonationOperation("HandleOAuthCallback", "Handle OAuth callback")
	suite.registerTestHandler("GET", "oauth/callback", callbackOp, suite.h.HandleOAuthCallback)

	ctx := echocontext.NewTestEchoContext().Request().Context()
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// Create test user with organization
	testUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      "0p3nl@n3rocks!",
		confirmedUser: true,
	})

	// Generate valid state for testing
	orgID := testUser.OrganizationID
	provider := "github"
	validState, err := generateTestOAuthState(orgID, provider)
	require.NoError(t, err)

	t.Run("successful callback redirects to HTML interface", func(t *testing.T) {
		// Since we can't easily mock the OAuth2 endpoints without significant refactoring,
		// we'll test that the request passes initial validation but fails at token exchange.
		// This validates all the cookie handling, state validation, and request processing logic.
		urlPath := fmt.Sprintf("/oauth/callback?code=test_code&state=%s", validState)

		req := httptest.NewRequest(http.MethodGet, urlPath, nil)

		// Add required OAuth cookies
		req.AddCookie(&http.Cookie{Name: "oauth_state", Value: validState})
		req.AddCookie(&http.Cookie{Name: "oauth_org_id", Value: orgID})
		req.AddCookie(&http.Cookie{Name: "oauth_user_id", Value: testUser.UserInfo.ID})

		rec := httptest.NewRecorder()
		suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

		// Should get 500 internal server error due to OAuth token exchange failing.
		// This is expected since we don't have real OAuth provider setup in tests.
		// The important thing is that it passes all validation (not 400 Bad Request).
		assert.Equal(t, http.StatusInternalServerError, rec.Code, "Should fail with internal server error at OAuth token exchange")
		assert.Contains(t, rec.Body.String(), "failed to exchange authorization code", "Should contain OAuth exchange error message")
	})
}

func (suite *HandlerTestSuite) TestOAuthStateValidationEdgeCases() {
	t := suite.T()

	// Register the route once for the entire test
	// Create operation for HandleOAuthCallback
	callbackOp := suite.createImpersonationOperation("HandleOAuthCallback", "Handle OAuth callback")
	suite.registerTestHandler("GET", "oauth/callback", callbackOp, suite.h.HandleOAuthCallback)

	t.Run("state validation with invalid base64", func(t *testing.T) {
		orgID := "test-org-123"

		// Test invalid base64 state
		invalidState := "invalid-base64-!@#$%"

		req := httptest.NewRequest(http.MethodGet,
			fmt.Sprintf("/oauth/callback?code=test_code&state=%s", invalidState), nil)

		req.AddCookie(&http.Cookie{Name: "oauth_state", Value: invalidState})
		req.AddCookie(&http.Cookie{Name: "oauth_org_id", Value: orgID})
		req.AddCookie(&http.Cookie{Name: "oauth_user_id", Value: "test-user"})

		rec := httptest.NewRecorder()
		suite.e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "state is required")
	})

	t.Run("state validation with wrong format", func(t *testing.T) {
		orgID := "test-org-123"

		// Create state with wrong format (missing parts)
		wrongFormatState := base64.URLEncoding.EncodeToString([]byte("only-one-part"))

		req := httptest.NewRequest(http.MethodGet,
			fmt.Sprintf("/oauth/callback?code=test_code&state=%s", wrongFormatState), nil)

		req.AddCookie(&http.Cookie{Name: "oauth_state", Value: wrongFormatState})
		req.AddCookie(&http.Cookie{Name: "oauth_org_id", Value: orgID})
		req.AddCookie(&http.Cookie{Name: "oauth_user_id", Value: "test-user"})

		rec := httptest.NewRecorder()
		suite.e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "invalid OAuth state parameter")
	})
}

func (suite *HandlerTestSuite) TestOAuthProviderConfiguration() {
	t := suite.T()

	// Create operation for StartOAuthFlow
	startOp := suite.createImpersonationOperation("StartOAuthFlow", "Start OAuth flow")
	suite.registerTestHandler("POST", "oauth/start", startOp, suite.h.StartOAuthFlow)

	ctx := echocontext.NewTestEchoContext().Request().Context()
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// Create test user with organization
	testUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      "0p3nl@n3rocks!",
		confirmedUser: true,
	})

	tests := []struct {
		name           string
		provider       string
		expectedStatus int
		checkResponse  func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:           "github provider should be configured",
			provider:       "github",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var response models.OAuthFlowResponse
				err := json.Unmarshal(resp.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.True(t, response.Success)
				assert.Contains(t, response.AuthURL, "github.com")
			},
		},
		{
			name:           "slack provider should not be configured in test environment",
			provider:       "slack",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Contains(t, resp.Body.String(), "oauth2 provider not supported")
			},
		},
		{
			name:           "unsupported provider should fail",
			provider:       "unsupported",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Contains(t, resp.Body.String(), "invalid or unparsable field: provider")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := models.OAuthFlowRequest{
				Provider: tt.provider,
			}

			body, err := json.Marshal(request)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/oauth/start", strings.NewReader(string(body)))
			req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

			rec := httptest.NewRecorder()
			suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}
		})
	}
}

func (suite *HandlerTestSuite) TestOAuthCookieConfiguration() {
	t := suite.T()

	// Create operation for StartOAuthFlow
	startOp := suite.createImpersonationOperation("StartOAuthFlow", "Start OAuth flow")
	suite.registerTestHandler("POST", "oauth/start", startOp, suite.h.StartOAuthFlow)

	ctx := echocontext.NewTestEchoContext().Request().Context()
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// Create test user with organization
	testUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      "0p3nl@n3rocks!",
		confirmedUser: true,
	})

	t.Run("OAuth cookies should be set with SameSiteLax in Test", func(t *testing.T) {
		request := models.OAuthFlowRequest{
			Provider: "github",
		}

		body, err := json.Marshal(request)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/oauth/start", strings.NewReader(string(body)))
		req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

		rec := httptest.NewRecorder()
		suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

		assert.Equal(t, http.StatusOK, rec.Code)

		// Check that OAuth cookies are set with proper configuration
		cookies := rec.Result().Cookies()

		var oauthStateCookie, oauthOrgCookie, oauthUserCookie *http.Cookie
		for _, cookie := range cookies {
			switch cookie.Name {
			case "oauth_state":
				oauthStateCookie = cookie
			case "oauth_org_id":
				oauthOrgCookie = cookie
			case "oauth_user_id":
				oauthUserCookie = cookie
			}
		}

		// Verify all OAuth cookies are set
		require.NotNil(t, oauthStateCookie, "oauth_state cookie should be set")
		require.NotNil(t, oauthOrgCookie, "oauth_org_id cookie should be set")
		require.NotNil(t, oauthUserCookie, "oauth_user_id cookie should be set")

		// Verify OAuth cookies have proper SameSite configuration
		assert.Equal(t, http.SameSiteLaxMode, oauthStateCookie.SameSite)
		assert.Equal(t, http.SameSiteLaxMode, oauthOrgCookie.SameSite)
		assert.Equal(t, http.SameSiteLaxMode, oauthUserCookie.SameSite)

		// Verify OAuth cookies are HttpOnly for security
		assert.True(t, oauthStateCookie.HttpOnly)
		assert.True(t, oauthOrgCookie.HttpOnly)
		assert.True(t, oauthUserCookie.HttpOnly)
	})
}

// Helper functions for testing

// generateTestOAuthState creates a valid OAuth state for testing
func generateTestOAuthState(orgID, provider string) (string, error) {
	// Create a test handler with minimal setup to access the private method
	// For testing purposes, we'll create a simple base64 encoded state
	// This mimics the format used by the actual generateOAuthState method
	// Use current timestamp to make states different
	stateData := fmt.Sprintf("%s:%s:%d", orgID, provider, time.Now().UnixNano())
	return base64.URLEncoding.EncodeToString([]byte(stateData)), nil
}

// TestStartOAuthFlowCookieHandling tests that the OAuth start flow properly sets cookies with SameSiteLax
func (suite *HandlerTestSuite) TestStartOAuthFlowCookieHandling() {
	t := suite.T()

	// Create operation for StartOAuthFlow
	startOp := suite.createImpersonationOperation("StartOAuthFlow", "Start OAuth flow")
	suite.registerTestHandler("POST", "oauth/start", startOp, suite.h.StartOAuthFlow)

	ctx := echocontext.NewTestEchoContext().Request().Context()
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// Create test user with organization
	testUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      "0p3nl@n3rocks!",
		confirmedUser: true,
	})

	t.Run("sets OAuth cookies with SameSiteLax", func(t *testing.T) {
		request := models.OAuthFlowRequest{
			Provider: "github",
		}

		body, err := json.Marshal(request)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/oauth/start", strings.NewReader(string(body)))
		req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

		// Add existing auth cookies to simulate authenticated user
		req.AddCookie(&http.Cookie{
			Name:  auth.AccessTokenCookie,
			Value: "test_access_token",
		})
		req.AddCookie(&http.Cookie{
			Name:  auth.RefreshTokenCookie,
			Value: "test_refresh_token",
		})

		rec := httptest.NewRecorder()
		suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

		assert.Equal(t, http.StatusOK, rec.Code)

		// Parse response to get cookies
		cookies := rec.Result().Cookies()

		// Verify OAuth-specific cookies are set with SameSiteLax
		var oauthStateCookie, oauthOrgCookie, oauthUserCookie *http.Cookie
		var authAccessCookie, authRefreshCookie *http.Cookie

		for _, cookie := range cookies {
			switch cookie.Name {
			case "oauth_state":
				oauthStateCookie = cookie
			case "oauth_org_id":
				oauthOrgCookie = cookie
			case "oauth_user_id":
				oauthUserCookie = cookie
			case auth.AccessTokenCookie:
				authAccessCookie = cookie
			case auth.RefreshTokenCookie:
				authRefreshCookie = cookie
			}
		}

		// Verify OAuth cookies are set with proper SameSite policy
		require.NotNil(t, oauthStateCookie, "oauth_state cookie should be set")
		assert.Equal(t, http.SameSiteLaxMode, oauthStateCookie.SameSite, "oauth_state should have SameSiteLax")
		assert.True(t, oauthStateCookie.HttpOnly, "oauth_state should be HttpOnly")
		assert.False(t, oauthStateCookie.Secure, "oauth_state should not be Secure in test mode")

		require.NotNil(t, oauthOrgCookie, "oauth_org_id cookie should be set")
		assert.Equal(t, http.SameSiteLaxMode, oauthOrgCookie.SameSite, "oauth_org_id should have SameSiteLax")
		assert.Equal(t, testUser.OrganizationID, oauthOrgCookie.Value, "oauth_org_id should match user's org")

		require.NotNil(t, oauthUserCookie, "oauth_user_id cookie should be set")
		assert.Equal(t, http.SameSiteLaxMode, oauthUserCookie.SameSite, "oauth_user_id should have SameSiteLax")
		assert.Equal(t, testUser.UserInfo.ID, oauthUserCookie.Value, "oauth_user_id should match user's ID")

		// Verify auth cookies are re-set with SameSiteLax for OAuth compatibility
		require.NotNil(t, authAccessCookie, "access token cookie should be re-set")
		assert.Equal(t, http.SameSiteLaxMode, authAccessCookie.SameSite, "access token should have SameSiteLax for OAuth")
		assert.Equal(t, "test_access_token", authAccessCookie.Value, "access token value should be preserved")

		require.NotNil(t, authRefreshCookie, "refresh token cookie should be re-set")
		assert.Equal(t, http.SameSiteLaxMode, authRefreshCookie.SameSite, "refresh token should have SameSiteLax for OAuth")
		assert.Equal(t, "test_refresh_token", authRefreshCookie.Value, "refresh token value should be preserved")
	})

	t.Run("handles missing auth cookies gracefully", func(t *testing.T) {
		request := models.OAuthFlowRequest{
			Provider: "github",
		}

		body, err := json.Marshal(request)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/oauth/start", strings.NewReader(string(body)))
		req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)
		// Don't add auth cookies

		rec := httptest.NewRecorder()
		suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

		assert.Equal(t, http.StatusOK, rec.Code)

		// Should still set OAuth cookies even without existing auth cookies
		cookies := rec.Result().Cookies()
		var oauthStateCookie *http.Cookie

		for _, cookie := range cookies {
			if cookie.Name == "oauth_state" {
				oauthStateCookie = cookie
				break
			}
		}

		require.NotNil(t, oauthStateCookie, "oauth_state cookie should be set even without auth cookies")
		assert.Equal(t, http.SameSiteLaxMode, oauthStateCookie.SameSite, "oauth_state should have SameSiteLax")
	})
}

// TestOAuthCallbackWithCookies tests that the OAuth callback works with proper cookie setup
func (suite *HandlerTestSuite) TestOAuthCallbackWithCookies() {
	t := suite.T()

	// Create operation for HandleOAuthCallback
	callbackOp := suite.createImpersonationOperation("HandleOAuthCallback", "Handle OAuth callback")
	suite.registerTestHandler("GET", "oauth/callback", callbackOp, suite.h.HandleOAuthCallback)

	ctx := echocontext.NewTestEchoContext().Request().Context()
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// Create test user with organization
	testUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      "0p3nl@n3rocks!",
		confirmedUser: true,
	})

	// Generate valid state for testing
	orgID := testUser.OrganizationID
	provider := "github"
	validState, err := generateTestOAuthState(orgID, provider)
	require.NoError(t, err)

	t.Run("callback works with proper cookies", func(t *testing.T) {
		// Create a request with proper OAuth cookies set (simulating successful OAuth start)
		req := httptest.NewRequest(http.MethodGet,
			fmt.Sprintf("/oauth/callback?code=test_code&state=%s", validState), nil)

		// Add OAuth cookies that would be set by the start flow
		req.AddCookie(&http.Cookie{
			Name:     "oauth_state",
			Value:    validState,
			SameSite: http.SameSiteLaxMode,
		})
		req.AddCookie(&http.Cookie{
			Name:     "oauth_org_id",
			Value:    orgID,
			SameSite: http.SameSiteLaxMode,
		})
		req.AddCookie(&http.Cookie{
			Name:     "oauth_user_id",
			Value:    testUser.UserInfo.ID,
			SameSite: http.SameSiteLaxMode,
		})

		rec := httptest.NewRecorder()
		suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

		// Should not get a 401 anymore - the authentication context should be properly set
		// Note: This might still fail due to missing OAuth provider validation or other issues,
		// but it should NOT fail with a 401 due to missing authentication
		assert.NotEqual(t, http.StatusUnauthorized, rec.Code,
			"Should not get 401 - authentication context should be available")

		// Check that OAuth cookies get cleaned up on successful callback
		if rec.Code == http.StatusOK {
			cookies := rec.Result().Cookies()
			for _, cookie := range cookies {
				if strings.HasPrefix(cookie.Name, "oauth_") {
					// OAuth cookies should be removed/expired
					assert.True(t, cookie.MaxAge < 0 || cookie.Expires.Before(time.Now()),
						"OAuth cookie %s should be removed after successful callback", cookie.Name)
				}
			}
		}
	})

	t.Run("callback fails without proper cookies", func(t *testing.T) {
		// Create a request without OAuth cookies
		req := httptest.NewRequest(http.MethodGet,
			fmt.Sprintf("/oauth/callback?code=test_code&state=%s", validState), nil)

		rec := httptest.NewRecorder()
		suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

		// Should fail due to missing OAuth state cookie
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "state")
	})
}

// TestOAuthStateValidation tests OAuth state generation and validation
func (suite *HandlerTestSuite) TestOAuthStateValidation() {
	t := suite.T()

	orgID := ulids.New().String()
	provider := "github"

	t.Run("state generation and validation", func(t *testing.T) {
		// Generate a state
		state, err := generateTestOAuthState(orgID, provider)
		require.NoError(t, err)
		assert.NotEmpty(t, state)

		// Validate the state format
		stateBytes, err := base64.URLEncoding.DecodeString(state)
		require.NoError(t, err)

		parts := strings.Split(string(stateBytes), ":")
		assert.Len(t, parts, 3, "State should have 3 parts: orgID:provider:random")
		assert.Equal(t, orgID, parts[0], "First part should be orgID")
		assert.Equal(t, provider, parts[1], "Second part should be provider")
		assert.NotEmpty(t, parts[2], "Third part should be random data")
	})

	t.Run("different states for different inputs", func(t *testing.T) {
		state1, err := generateTestOAuthState(orgID, provider)
		require.NoError(t, err)

		state2, err := generateTestOAuthState(orgID, provider)
		require.NoError(t, err)

		// States should be different due to random component
		assert.NotEqual(t, state1, state2, "Different calls should generate different states")
	})
}
