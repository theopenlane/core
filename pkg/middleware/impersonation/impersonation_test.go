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
		AccessDuration:  time.Hour,
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

func TestMiddlewareProcessNoToken(t *testing.T) {
	e := echo.New()
	tokenManager := setupTestTokenManager(t)
	middleware := New(tokenManager)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	nextCalled := false
	handler := func(c echo.Context) error {
		nextCalled = true
		return c.String(http.StatusOK, "OK")
	}

	err := middleware.Process(handler)(c)

	assert.NoError(t, err)
	assert.True(t, nextCalled)
	_, ok := auth.CallerFromContext(c.Request().Context())
	assert.False(t, ok)
}

func TestMiddlewareProcessInvalidToken(t *testing.T) {
	e := echo.New()
	tokenManager := setupTestTokenManager(t)
	middleware := New(tokenManager)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Impersonation eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	nextCalled := false
	handler := func(c echo.Context) error {
		nextCalled = true
		return c.String(http.StatusOK, "OK")
	}

	err := middleware.Process(handler)(c)

	assert.Error(t, err)
	assert.False(t, nextCalled)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
}

func TestMiddlewareProcessValidToken(t *testing.T) {
	e := echo.New()
	tokenManager := setupTestTokenManager(t)

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

	middleware := New(tokenManager)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Impersonation "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	nextCalled := false
	handler := func(c echo.Context) error {
		nextCalled = true

		caller, ok := auth.CallerFromContext(c.Request().Context())
		assert.True(t, ok)
		assert.NotNil(t, caller)
		assert.True(t, caller.IsImpersonated())
		assert.Equal(t, "target-user-123", caller.SubjectID)
		assert.Equal(t, "target@example.com", caller.SubjectEmail)
		assert.Equal(t, "admin-123", caller.Impersonation.ImpersonatorID)
		assert.Equal(t, []string{"read", "debug"}, caller.Impersonation.Scopes)

		return c.String(http.StatusOK, "OK")
	}

	err = middleware.Process(handler)(c)

	assert.NoError(t, err)
	assert.True(t, nextCalled)
}

func TestCreateImpersonatedCaller(t *testing.T) {
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

	caller, err := middleware.createImpersonatedCaller(claims)

	assert.NoError(t, err)
	assert.NotNil(t, caller)
	assert.Equal(t, "user-123", caller.SubjectID)
	assert.Equal(t, "user@example.com", caller.SubjectEmail)
	assert.Equal(t, "org-456", caller.OrganizationID)
	assert.Equal(t, auth.JWTAuthentication, caller.AuthenticationType)
	assert.NotNil(t, caller.Impersonation)
	assert.Equal(t, auth.ImpersonationType("support"), caller.Impersonation.Type)
	assert.Equal(t, "admin-789", caller.Impersonation.ImpersonatorID)
	assert.Equal(t, "session-123", caller.Impersonation.SessionID)
	assert.Equal(t, []string{"read", "debug"}, caller.Impersonation.Scopes)
}

func TestRequireImpersonationScope(t *testing.T) {
	tests := []struct {
		name           string
		requiredScope  string
		setupContext   func() context.Context
		expectedStatus int
	}{
		{
			name:           "not impersonated passes through",
			requiredScope:  "admin:write",
			setupContext:   context.Background,
			expectedStatus: http.StatusOK,
		},
		{
			name:          "has required scope allows access",
			requiredScope: "read",
			setupContext: func() context.Context {
				return auth.WithCaller(context.Background(), &auth.Caller{
					SubjectID: "user123",
					Impersonation: &auth.ImpersonationContext{
						Scopes:    []string{"read", "debug"},
						ExpiresAt: time.Now().Add(time.Hour),
					},
				})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "missing required scope denies access",
			requiredScope: "write",
			setupContext: func() context.Context {
				return auth.WithCaller(context.Background(), &auth.Caller{
					SubjectID: "user123",
					Impersonation: &auth.ImpersonationContext{
						Scopes:    []string{"read", "debug"},
						ExpiresAt: time.Now().Add(time.Hour),
					},
				})
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			middleware := RequireImpersonationScope(tt.requiredScope)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req = req.WithContext(tt.setupContext())
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := func(c echo.Context) error {
				return c.String(http.StatusOK, "OK")
			}

			err := middleware(handler)(c)
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
			name:           "not impersonated allows access",
			setupContext:   context.Background,
			expectedStatus: http.StatusOK,
		},
		{
			name: "impersonated blocks access",
			setupContext: func() context.Context {
				return auth.WithCaller(context.Background(), &auth.Caller{
					SubjectID: "user123",
					Impersonation: &auth.ImpersonationContext{
						Type: auth.SupportImpersonation,
					},
				})
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			middleware := BlockImpersonation()

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req = req.WithContext(tt.setupContext())
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := func(c echo.Context) error {
				return c.String(http.StatusOK, "OK")
			}

			err := middleware(handler)(c)
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
			name:           "not impersonated allows access",
			allowedTypes:   []auth.ImpersonationType{auth.SupportImpersonation},
			setupContext:   context.Background,
			expectedStatus: http.StatusOK,
		},
		{
			name:         "allowed type permits access",
			allowedTypes: []auth.ImpersonationType{auth.SupportImpersonation, auth.AdminImpersonation},
			setupContext: func() context.Context {
				return auth.WithCaller(context.Background(), &auth.Caller{
					SubjectID: "user123",
					Impersonation: &auth.ImpersonationContext{
						Type: auth.SupportImpersonation,
					},
				})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "disallowed type denies access",
			allowedTypes: []auth.ImpersonationType{auth.SupportImpersonation},
			setupContext: func() context.Context {
				return auth.WithCaller(context.Background(), &auth.Caller{
					SubjectID: "user123",
					Impersonation: &auth.ImpersonationContext{
						Type: auth.AdminImpersonation,
					},
				})
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			middleware := AllowOnlyImpersonationType(tt.allowedTypes...)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req = req.WithContext(tt.setupContext())
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := func(c echo.Context) error {
				return c.String(http.StatusOK, "OK")
			}

			err := middleware(handler)(c)
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
		name         string
		setupContext func() context.Context
		setupHeaders func(*http.Request)
		checkContext func(*testing.T, echo.Context)
	}{
		{
			name:         "no authenticated user passes through",
			setupContext: context.Background,
		},
		{
			name: "system admin with headers switches context",
			setupContext: func() context.Context {
				return auth.WithCaller(context.Background(), &auth.Caller{
					SubjectID:      "admin123",
					SubjectEmail:   "admin@example.com",
					OrganizationID: "org-admin",
					Capabilities:   auth.CapSystemAdmin,
				})
			},
			setupHeaders: func(r *http.Request) {
				r.Header.Set("X-User-ID", "target-user-456")
				r.Header.Set("X-Organization-ID", "target-org-789")
			},
			checkContext: func(t *testing.T, c echo.Context) {
				user, ok := auth.CallerFromContext(c.Request().Context())
				assert.True(t, ok)
				assert.Equal(t, "target-user-456", user.SubjectID)
				assert.Equal(t, "target-org-789", user.OrganizationID)
				assert.False(t, user.Has(auth.CapSystemAdmin))

				adminUser, ok := auth.OriginalSystemAdminCallerFromContext(c.Request().Context())
				assert.True(t, ok)
				assert.Equal(t, "admin123", adminUser.SubjectID)
				assert.True(t, adminUser.Has(auth.CapSystemAdmin))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			middleware := SystemAdminUserContextMiddleware()

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req = req.WithContext(tt.setupContext())
			if tt.setupHeaders != nil {
				tt.setupHeaders(req)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := func(c echo.Context) error {
				return c.String(http.StatusOK, "OK")
			}

			err := middleware(handler)(c)
			assert.NoError(t, err)

			if tt.checkContext != nil {
				tt.checkContext(t, c)
			}
		})
	}
}

func TestLogImpersonationAccess(t *testing.T) {
	middleware := &Middleware{}

	claims := &tokens.ImpersonationClaims{
		ImpersonatorID: "admin-123",
		UserID:         "user-456",
	}

	middleware.logImpersonationAccess(claims, nil)
}
