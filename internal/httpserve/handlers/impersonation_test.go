package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/echox/middleware/echocontext"

	"github.com/theopenlane/common/enums"
	models "github.com/theopenlane/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

func (suite *HandlerTestSuite) TestStartImpersonation() {
	t := suite.T()

	// Register test handler
	suite.registerTestHandler("POST", "impersonation/start", suite.startImpersonationOp, suite.h.StartImpersonation)

	ec := echocontext.NewTestEchoContext().Request().Context()
	ctx := privacy.DecisionContext(ec, privacy.Allow)

	// Create test users - admin and target user
	adminID := ulids.New().String()
	targetID := ulids.New().String()
	orgID := ulids.New().String()

	adminUserSetting := suite.db.UserSetting.Create().
		SetEmailConfirmed(true).
		SaveX(ctx)

	adminUser := suite.db.User.Create().
		SetID(adminID).
		SetFirstName("Admin").
		SetLastName("User").
		SetEmail("admin-" + adminID + "@example.com").
		SetPassword("SecurePassword123!").
		SetSetting(adminUserSetting).
		SetRole(enums.RoleAdmin).
		SetLastLoginProvider(enums.AuthProviderCredentials).
		SetLastSeen(time.Now()).
		SetSub(adminID).
		SaveX(ctx)

	targetUserSetting := suite.db.UserSetting.Create().
		SetEmailConfirmed(true).
		SaveX(ctx)

	targetUser := suite.db.User.Create().
		SetID(targetID).
		SetFirstName("Target").
		SetLastName("User").
		SetEmail("target-" + targetID + "@example.com").
		SetPassword("SecurePassword123!").
		SetSetting(targetUserSetting).
		SetRole(enums.RoleUser).
		SetLastLoginProvider(enums.AuthProviderCredentials).
		SetLastSeen(time.Now()).
		SetSub(targetID).
		SaveX(ctx)

	testCases := []struct {
		name           string
		request        models.StartImpersonationRequest
		setupContext   func() context.Context
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful support impersonation",
			request: models.StartImpersonationRequest{
				TargetUserID: targetUser.ID,
				Type:         "support",
				Reason:       "debugging issue for customer support",
				Duration:     intPtr(2),
				Scopes:       []string{"read", "debug"},
			},
			setupContext: func() context.Context {
				ctx := context.Background()
				user := &auth.AuthenticatedUser{
					SubjectID:      adminUser.ID,
					SubjectEmail:   adminUser.Email,
					OrganizationID: orgID,
					IsSystemAdmin:  true, // System admin can perform impersonation
				}
				return auth.WithAuthenticatedUser(ctx, user)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.StartImpersonationReply
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.True(t, response.Success)
				assert.NotEmpty(t, response.Token)
				assert.NotEmpty(t, response.SessionID)
				assert.Equal(t, "Impersonation session started successfully", response.Message)
			},
		},
		{
			name: "invalid request - missing target user ID",
			request: models.StartImpersonationRequest{
				Type:   "support",
				Reason: "debugging user account issues",
			},
			setupContext: func() context.Context {
				ctx := context.Background()
				user := &auth.AuthenticatedUser{
					SubjectID:     adminUser.ID,
					IsSystemAdmin: true,
				}
				return auth.WithAuthenticatedUser(ctx, user)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "no authenticated user",
			request: models.StartImpersonationRequest{
				TargetUserID: targetUser.ID,
				Type:         "support",
				Reason:       "debugging user account issues",
			},
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "insufficient permissions - not system admin",
			request: models.StartImpersonationRequest{
				TargetUserID: targetUser.ID,
				Type:         "support",
				Reason:       "debugging user account issues",
			},
			setupContext: func() context.Context {
				ctx := context.Background()
				user := &auth.AuthenticatedUser{
					SubjectID:     "user-456",
					IsSystemAdmin: false,
				}
				return auth.WithAuthenticatedUser(ctx, user)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "target user not found",
			request: models.StartImpersonationRequest{
				TargetUserID: "nonexistent-user",
				Type:         "support",
				Reason:       "debugging user account issues",
			},
			setupContext: func() context.Context {
				ctx := context.Background()
				user := &auth.AuthenticatedUser{
					SubjectID:     adminUser.ID,
					IsSystemAdmin: true,
				}
				return auth.WithAuthenticatedUser(ctx, user)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal request
			reqBody, err := json.Marshal(tt.request)
			require.NoError(t, err)

			// Create HTTP request
			req := httptest.NewRequest(http.MethodPost, "/impersonation/start", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Set up request context with both auth and privacy context
			requestCtx := tt.setupContext()
			requestCtx = privacy.DecisionContext(requestCtx, privacy.Allow)
			req = req.WithContext(requestCtx)

			rec := httptest.NewRecorder()

			// Using the ServeHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(rec, req)

			// Check status
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Check response if provided
			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}
		})
	}
}

func (suite *HandlerTestSuite) TestEndImpersonation() {
	t := suite.T()

	// Register test handler
	suite.registerTestHandler("POST", "impersonation/end", suite.endImpersonationOp, suite.h.EndImpersonation)

	testCases := []struct {
		name           string
		request        models.EndImpersonationRequest
		setupContext   func() context.Context
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful impersonation end",
			request: models.EndImpersonationRequest{
				SessionID: "session-123",
				Reason:    "task completed",
			},
			setupContext: func() context.Context {
				ctx := context.Background()
				impUser := &auth.ImpersonatedUser{
					AuthenticatedUser: &auth.AuthenticatedUser{
						SubjectID: "user-123",
					},
					ImpersonationContext: &auth.ImpersonationContext{
						SessionID:         "session-123",
						ImpersonatorID:    "admin-456",
						TargetUserID:      "user-123",
						ImpersonatorEmail: "admin@example.com",
						TargetUserEmail:   "user@example.com",
						Type:              auth.SupportImpersonation,
						Scopes:            []string{"read", "debug"},
					},
					OriginalUser: &auth.AuthenticatedUser{
						SubjectID: "admin-456",
					},
				}
				return auth.WithImpersonatedUser(ctx, impUser)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.EndImpersonationReply
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.True(t, response.Success)
				assert.Equal(t, "Impersonation session ended successfully", response.Message)
			},
		},
		{
			name: "invalid request - missing session ID",
			request: models.EndImpersonationRequest{
				Reason: "task completed",
			},
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "no active impersonation session",
			request: models.EndImpersonationRequest{
				SessionID: "session-123",
			},
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "session ID mismatch",
			request: models.EndImpersonationRequest{
				SessionID: "wrong-session-456",
			},
			setupContext: func() context.Context {
				ctx := context.Background()
				impUser := &auth.ImpersonatedUser{
					AuthenticatedUser: &auth.AuthenticatedUser{
						SubjectID: "user-123",
					},
					ImpersonationContext: &auth.ImpersonationContext{
						SessionID: "session-123",
					},
				}
				return auth.WithImpersonatedUser(ctx, impUser)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal request
			reqBody, err := json.Marshal(tt.request)
			require.NoError(t, err)

			// Create HTTP request
			req := httptest.NewRequest(http.MethodPost, "/impersonation/end", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Set up request context with both auth and privacy context
			requestCtx := tt.setupContext()
			requestCtx = privacy.DecisionContext(requestCtx, privacy.Allow)
			req = req.WithContext(requestCtx)

			rec := httptest.NewRecorder()

			// Using the ServeHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(rec, req)

			// Check status
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Check response if provided
			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}
		})
	}
}

func (suite *HandlerTestSuite) TestExtractSessionIDFromToken() {
	t := suite.T()

	// Register test handler
	suite.registerTestHandler("POST", "impersonation/start", suite.startImpersonationOp, suite.h.StartImpersonation)

	// Create a valid impersonation token using the real TokenManager
	opts := tokens.CreateImpersonationTokenOptions{
		ImpersonatorID:    ulids.New().String(),
		ImpersonatorEmail: "admin@example.com",
		TargetUserID:      ulids.New().String(),
		TargetUserEmail:   "user@example.com",
		OrganizationID:    ulids.New().String(),
		Type:              "support",
		Reason:            "testing",
		Duration:          time.Hour,
		Scopes:            []string{"read"},
	}

	_, err := suite.h.TokenManager.CreateImpersonationToken(context.Background(), opts)
	require.NoError(t, err)

	// Test valid token by creating a simple test handler to call extractSessionIDFromToken
	// Since it's a private method, we test it indirectly through StartImpersonation
	// which internally calls extractSessionIDFromToken

	// Create test context and request that should succeed
	ec := echocontext.NewTestEchoContext().Request().Context()
	ctx := privacy.DecisionContext(ec, privacy.Allow)

	// Create admin and target users
	adminID := ulids.New().String()
	targetID := ulids.New().String()

	adminUserSetting := suite.db.UserSetting.Create().
		SetEmailConfirmed(true).
		SaveX(ctx)

	adminUser := suite.db.User.Create().
		SetID(adminID).
		SetFirstName("Admin").
		SetLastName("User").
		SetEmail("admin-" + adminID + "@example.com").
		SetPassword("SecurePassword123!").
		SetSetting(adminUserSetting).
		SetRole(enums.RoleAdmin).
		SetLastLoginProvider(enums.AuthProviderCredentials).
		SetLastSeen(time.Now()).
		SetSub(adminID).
		SaveX(ctx)

	targetUserSetting := suite.db.UserSetting.Create().
		SetEmailConfirmed(true).
		SaveX(ctx)

	targetUser := suite.db.User.Create().
		SetID(targetID).
		SetFirstName("Target").
		SetLastName("User").
		SetEmail("target-" + targetID + "@example.com").
		SetPassword("SecurePassword123!").
		SetSetting(targetUserSetting).
		SetRole(enums.RoleUser).
		SetLastLoginProvider(enums.AuthProviderCredentials).
		SetLastSeen(time.Now()).
		SetSub(targetID).
		SaveX(ctx)

	// Test that token creation and session ID extraction works
	request := models.StartImpersonationRequest{
		TargetUserID: targetUser.ID,
		Type:         "support",
		Reason:       "testing session ID extraction",
		Duration:     intPtr(1),
		Scopes:       []string{"read"},
	}

	reqBody, err := json.Marshal(request)
	require.NoError(t, err)

	testCtx := context.Background()
	user := &auth.AuthenticatedUser{
		SubjectID:      adminUser.ID,
		SubjectEmail:   adminUser.Email,
		OrganizationID: ulids.New().String(),
		IsSystemAdmin:  true,
	}
	testCtx = auth.WithAuthenticatedUser(testCtx, user)

	req := httptest.NewRequest(http.MethodPost, "/impersonation/start", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Set up request context with both auth and privacy context
	testCtx = privacy.DecisionContext(testCtx, privacy.Allow)
	req = req.WithContext(testCtx)

	rec := httptest.NewRecorder()

	// Using the ServeHTTP on echo will trigger the router and middleware
	suite.e.ServeHTTP(rec, req)

	// Execute handler - this will internally test extractSessionIDFromToken
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify the response includes a session ID (which means extractSessionIDFromToken worked)
	var response models.StartImpersonationReply
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotEmpty(t, response.SessionID)
}

// Helper functions
func intPtr(i int) *int {
	return &i
}
