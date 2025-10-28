package impersonation

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
)

var testKey ed25519.PrivateKey

func init() {
	var err error
	_, testKey, err = ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
}

func setupTestTokenManager(t *testing.T) *tokens.TokenManager {
	conf := tokens.Config{
		Audience:        "https://api.example.com",
		Issuer:          "https://auth.example.com",
		AccessDuration:  1 * time.Hour,
		RefreshDuration: 24 * time.Hour,
		RefreshOverlap:  -15 * time.Minute,
	}

	tm, err := tokens.NewWithKey(testKey, conf)
	assert.NoError(t, err)

	return tm
}

func TestNew(t *testing.T) {
	tokenManager := setupTestTokenManager(t)
	middleware := New(tokenManager)

	assert.NotNil(t, middleware)
	assert.Equal(t, tokenManager, middleware.tokenManager)
}

func TestMiddleware_Process_NoToken(t *testing.T) {
	// Create echo instance
	e := echo.New()

	// Create token manager
	tokenManager := setupTestTokenManager(t)

	// Create middleware
	middleware := New(tokenManager)

	// Create test request without impersonation header
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Track if next handler was called
	nextCalled := false
	handler := func(c echo.Context) error {
		nextCalled = true
		return c.String(http.StatusOK, "OK")
	}

	// Execute middleware
	err := middleware.Process(handler)(c)

	// Should pass through without error
	assert.NoError(t, err)
	assert.True(t, nextCalled)

	// Should not have impersonated user in context
	_, ok := auth.ImpersonatedUserFromContext(c.Request().Context())
	assert.False(t, ok)
}

func TestMiddleware_Process_MalformedHeader(t *testing.T) {
	// Create echo instance
	e := echo.New()

	// Create token manager
	tokenManager := setupTestTokenManager(t)

	// Create middleware
	middleware := New(tokenManager)

	// Create test request with malformed impersonation header (missing token part)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "invalid-format")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Track if next handler was called
	nextCalled := false
	handler := func(c echo.Context) error {
		nextCalled = true
		return c.String(http.StatusOK, "OK")
	}

	// Execute middleware
	err := middleware.Process(handler)(c)

	// Should pass through without error (auth.GetImpersonationToken returns error for malformed header)
	assert.NoError(t, err)
	assert.True(t, nextCalled)
}

func TestMiddleware_Process_InvalidToken(t *testing.T) {
	// Create echo instance
	e := echo.New()

	// Create token manager
	tokenManager := setupTestTokenManager(t)

	// Create middleware
	middleware := New(tokenManager)

	// Create test request with properly formatted but invalid impersonation token
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Impersonation eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Handler should not be called
	nextCalled := false
	handler := func(c echo.Context) error {
		nextCalled = true
		return c.String(http.StatusOK, "OK")
	}

	// Execute middleware
	err := middleware.Process(handler)(c)

	// Should return error
	assert.Error(t, err)
	assert.False(t, nextCalled)

	// Check it's an HTTP error with 401 status
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
}

func TestMiddleware_Process_ValidToken(t *testing.T) {
	// Create echo instance
	e := echo.New()

	// Create token manager
	tokenManager := setupTestTokenManager(t)

	// Create a valid impersonation token
	ctx := context.Background()
	opts := tokens.CreateImpersonationTokenOptions{
		ImpersonatorID:    "admin-123",
		ImpersonatorEmail: "admin@example.com",
		TargetUserID:      "target-user-123",
		TargetUserEmail:   "target@example.com",
		OrganizationID:    "org-123",
		Type:              "support",
		Reason:            "debugging issue",
		Scopes:            []string{"read", "debug"},
		Duration:          time.Hour,
	}

	token, err := tokenManager.CreateImpersonationToken(ctx, opts)
	assert.NoError(t, err)

	// Create middleware
	middleware := New(tokenManager)

	// Create test request with valid impersonation token
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Impersonation "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Track if next handler was called
	nextCalled := false
	handler := func(c echo.Context) error {
		nextCalled = true

		// Verify impersonated user is in context
		impUser, ok := auth.ImpersonatedUserFromContext(c.Request().Context())
		assert.True(t, ok, "should have impersonated user in context")
		assert.Equal(t, "target-user-123", impUser.SubjectID)
		assert.Equal(t, "target@example.com", impUser.SubjectEmail)
		assert.Equal(t, "admin-123", impUser.ImpersonationContext.ImpersonatorID)
		assert.Equal(t, []string{"read", "debug"}, impUser.ImpersonationContext.Scopes)

		return c.String(http.StatusOK, "OK")
	}

	// Execute middleware
	err = middleware.Process(handler)(c)

	// Should pass through without error
	assert.NoError(t, err)
	assert.True(t, nextCalled)
}

func TestCreateImpersonatedUser(t *testing.T) {
	middleware := &Middleware{}

	claims := &tokens.ImpersonationClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		UserID:            "user-123",
		TargetUserEmail:   "user@example.com",
		OrgID:             "org-456",
		ImpersonatorID:    "admin-789",
		ImpersonatorEmail: "admin@example.com",
		Type:              "support",
		Reason:            "investigating issue",
		SessionID:         "session-123",
		Scopes:            []string{"read", "debug"},
	}

	impUser, err := middleware.createImpersonatedUser(claims)

	assert.NoError(t, err)
	assert.NotNil(t, impUser)

	// Check target user
	assert.Equal(t, "user-123", impUser.SubjectID)
	assert.Equal(t, "user@example.com", impUser.SubjectEmail)
	assert.Equal(t, "org-456", impUser.OrganizationID)
	assert.Equal(t, auth.JWTAuthentication, impUser.AuthenticationType)

	// Check original user
	assert.Equal(t, "admin-789", impUser.OriginalUser.SubjectID)
	assert.Equal(t, "admin@example.com", impUser.OriginalUser.SubjectEmail)
	assert.Equal(t, "org-456", impUser.OriginalUser.OrganizationID)

	// Check impersonation context
	assert.Equal(t, auth.ImpersonationType("support"), impUser.ImpersonationContext.Type)
	assert.Equal(t, "admin-789", impUser.ImpersonationContext.ImpersonatorID)
	assert.Equal(t, "session-123", impUser.ImpersonationContext.SessionID)
	assert.Equal(t, []string{"read", "debug"}, impUser.ImpersonationContext.Scopes)
}

func TestRequireImpersonationScope(t *testing.T) {
	tests := []struct {
		name           string
		requiredScope  string
		setupContext   func() context.Context
		expectedStatus int
	}{
		{
			name:          "not impersonated - passes through",
			requiredScope: "admin:write",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "has required scope - allows access",
			requiredScope: "read",
			setupContext: func() context.Context {
				ctx := context.Background()
				impUser := &auth.ImpersonatedUser{
					AuthenticatedUser: &auth.AuthenticatedUser{
						SubjectID: "user123",
					},
					ImpersonationContext: &auth.ImpersonationContext{
						Scopes:    []string{"read", "debug"},
						ExpiresAt: time.Now().Add(time.Hour),
					},
				}
				return auth.WithImpersonatedUser(ctx, impUser)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "missing required scope - denies access",
			requiredScope: "write",
			setupContext: func() context.Context {
				ctx := context.Background()
				impUser := &auth.ImpersonatedUser{
					AuthenticatedUser: &auth.AuthenticatedUser{
						SubjectID: "user123",
					},
					ImpersonationContext: &auth.ImpersonationContext{
						Scopes:    []string{"read", "debug"},
						ExpiresAt: time.Now().Add(time.Hour),
					},
				}
				return auth.WithImpersonatedUser(ctx, impUser)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:          "wildcard scope allows all",
			requiredScope: "admin:write",
			setupContext: func() context.Context {
				ctx := context.Background()
				impUser := &auth.ImpersonatedUser{
					AuthenticatedUser: &auth.AuthenticatedUser{
						SubjectID: "user123",
					},
					ImpersonationContext: &auth.ImpersonationContext{
						Scopes:    []string{"*"},
						ExpiresAt: time.Now().Add(time.Hour),
					},
				}
				return auth.WithImpersonatedUser(ctx, impUser)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()

			// Create middleware
			middleware := RequireImpersonationScope(tt.requiredScope)

			// Create request with context
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req = req.WithContext(tt.setupContext())

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Handler that returns OK
			handler := func(c echo.Context) error {
				return c.String(http.StatusOK, "OK")
			}

			// Execute middleware
			err := middleware(handler)(c)

			// Check results
			if tt.expectedStatus == http.StatusOK {
				assert.NoError(t, err)
			} else {
				httpErr, ok := err.(*echo.HTTPError)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedStatus, httpErr.Code)
			}
		})
	}
}

func TestBlockImpersonation(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func() context.Context
		expectedStatus int
	}{
		{
			name: "not impersonated - allows access",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "impersonated - blocks access",
			setupContext: func() context.Context {
				ctx := context.Background()
				impUser := &auth.ImpersonatedUser{
					AuthenticatedUser: &auth.AuthenticatedUser{
						SubjectID: "user123",
					},
					ImpersonationContext: &auth.ImpersonationContext{
						Type: auth.SupportImpersonation,
					},
				}
				return auth.WithImpersonatedUser(ctx, impUser)
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()

			// Create middleware
			middleware := BlockImpersonation()

			// Create request with context
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req = req.WithContext(tt.setupContext())

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Handler that returns OK
			handler := func(c echo.Context) error {
				return c.String(http.StatusOK, "OK")
			}

			// Execute middleware
			err := middleware(handler)(c)

			// Check results
			if tt.expectedStatus == http.StatusOK {
				assert.NoError(t, err)
			} else {
				httpErr, ok := err.(*echo.HTTPError)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedStatus, httpErr.Code)
			}
		})
	}
}

func TestAllowOnlyImpersonationType(t *testing.T) {
	tests := []struct {
		name           string
		allowedTypes   []auth.ImpersonationType
		setupContext   func() context.Context
		expectedStatus int
	}{
		{
			name:         "not impersonated - allows access",
			allowedTypes: []auth.ImpersonationType{auth.SupportImpersonation},
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "allowed type - permits access",
			allowedTypes: []auth.ImpersonationType{auth.SupportImpersonation, auth.AdminImpersonation},
			setupContext: func() context.Context {
				ctx := context.Background()
				impUser := &auth.ImpersonatedUser{
					AuthenticatedUser: &auth.AuthenticatedUser{
						SubjectID: "user123",
					},
					ImpersonationContext: &auth.ImpersonationContext{
						Type: auth.SupportImpersonation,
					},
				}
				return auth.WithImpersonatedUser(ctx, impUser)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "disallowed type - denies access",
			allowedTypes: []auth.ImpersonationType{auth.SupportImpersonation},
			setupContext: func() context.Context {
				ctx := context.Background()
				impUser := &auth.ImpersonatedUser{
					AuthenticatedUser: &auth.AuthenticatedUser{
						SubjectID: "user123",
					},
					ImpersonationContext: &auth.ImpersonationContext{
						Type: auth.AdminImpersonation,
					},
				}
				return auth.WithImpersonatedUser(ctx, impUser)
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()

			// Create middleware
			middleware := AllowOnlyImpersonationType(tt.allowedTypes...)

			// Create request with context
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req = req.WithContext(tt.setupContext())

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Handler that returns OK
			handler := func(c echo.Context) error {
				return c.String(http.StatusOK, "OK")
			}

			// Execute middleware
			err := middleware(handler)(c)

			// Check results
			if tt.expectedStatus == http.StatusOK {
				assert.NoError(t, err)
			} else {
				httpErr, ok := err.(*echo.HTTPError)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedStatus, httpErr.Code)
			}
		})
	}
}

func TestSystemAdminUserContextMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func() context.Context
		setupHeaders   func(*http.Request)
		expectedStatus int
		checkContext   func(*testing.T, echo.Context)
	}{
		{
			name: "no authenticated user - passes through",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "non-admin user - passes through",
			setupContext: func() context.Context {
				ctx := context.Background()
				user := &auth.AuthenticatedUser{
					SubjectID:     "user123",
					IsSystemAdmin: false,
				}
				return auth.WithAuthenticatedUser(ctx, user)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "system admin without headers - passes through",
			setupContext: func() context.Context {
				ctx := context.Background()
				user := &auth.AuthenticatedUser{
					SubjectID:     "admin123",
					IsSystemAdmin: true,
				}
				return auth.WithAuthenticatedUser(ctx, user)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "system admin with user context headers - switches context",
			setupContext: func() context.Context {
				ctx := context.Background()
				user := &auth.AuthenticatedUser{
					SubjectID:      "admin123",
					SubjectEmail:   "admin@example.com",
					OrganizationID: "org-admin",
					IsSystemAdmin:  true,
				}
				return auth.WithAuthenticatedUser(ctx, user)
			},
			setupHeaders: func(r *http.Request) {
				r.Header.Set("X-User-ID", "target-user-456")
				r.Header.Set("X-Organization-ID", "target-org-789")
			},
			expectedStatus: http.StatusOK,
			checkContext: func(t *testing.T, c echo.Context) {
				// Check that the user context was switched
				user, ok := auth.AuthenticatedUserFromContext(c.Request().Context())
				assert.True(t, ok)
				assert.Equal(t, "target-user-456", user.SubjectID)
				assert.Equal(t, "target-org-789", user.OrganizationID)
				assert.False(t, user.IsSystemAdmin) // Target user should not inherit admin status

				// Check that original admin is stored
				adminUser, ok := auth.SystemAdminFromContext(c.Request().Context())
				assert.True(t, ok)
				assert.Equal(t, "admin123", adminUser.SubjectID)
				assert.True(t, adminUser.IsSystemAdmin)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()

			// Create middleware
			middleware := SystemAdminUserContextMiddleware()

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req = req.WithContext(tt.setupContext())

			if tt.setupHeaders != nil {
				tt.setupHeaders(req)
			}

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Handler that returns OK
			handler := func(c echo.Context) error {
				return c.String(http.StatusOK, "OK")
			}

			// Execute middleware
			err := middleware(handler)(c)

			// Check results
			assert.NoError(t, err)

			// Check context if provided
			if tt.checkContext != nil {
				tt.checkContext(t, c)
			}
		})
	}
}

func TestLogImpersonationAccess(t *testing.T) {
	// This is a simple logging function, just ensure it doesn't panic
	middleware := &Middleware{}

	claims := &tokens.ImpersonationClaims{
		ImpersonatorID: "admin-123",
		UserID:         "user-456",
	}

	// Should not panic
	middleware.logImpersonationAccess(claims, nil)
}
