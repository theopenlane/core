package openapi

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/theopenlane/utils/rout"
)

func TestOAuthFlowRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request OAuthFlowRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request - github",
			request: OAuthFlowRequest{
				Provider: "github",
			},
			wantErr: false,
		},
		{
			name: "valid request - github with scopes",
			request: OAuthFlowRequest{
				Provider: "github",
				Scopes:   []string{"repo", "gist"},
			},
			wantErr: false,
		},
		{
			name: "valid request - slack",
			request: OAuthFlowRequest{
				Provider: "slack",
			},
			wantErr: false,
		},
		{
			name: "valid request - with redirect URI",
			request: OAuthFlowRequest{
				Provider:    "github",
				RedirectURI: "https://app.example.com/callback",
			},
			wantErr: false,
		},
		{
			name: "empty provider",
			request: OAuthFlowRequest{
				Provider: "",
			},
			wantErr: true,
			errMsg:  "provider is required",
		},
		{
			name: "whitespace only provider",
			request: OAuthFlowRequest{
				Provider: "   ",
			},
			wantErr: true,
			errMsg:  "provider is required",
		},
		{
			name: "empty scopes should be filtered",
			request: OAuthFlowRequest{
				Provider: "github",
				Scopes:   []string{"repo", "", "gist", "   "},
			},
			wantErr: false,
		},
		{
			name: "case insensitive provider",
			request: OAuthFlowRequest{
				Provider: "GITHUB",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				// Check that provider is normalized to lowercase
				assert.Equal(t, tt.request.Provider, strings.ToLower(tt.request.Provider))
			}
		})
	}
}

func TestOAuthCallbackRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request OAuthCallbackRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: OAuthCallbackRequest{
				Provider: "github",
				Code:     "4/0AQlEz8xY...",
				State:    "eyJvcmdJRCI6IjAxSE...",
			},
			wantErr: false,
		},
		{
			name: "empty provider",
			request: OAuthCallbackRequest{
				Provider: "",
				Code:     "4/0AQlEz8xY...",
				State:    "eyJvcmdJRCI6IjAxSE...",
			},
			wantErr: false,
		},
		{
			name: "empty code",
			request: OAuthCallbackRequest{
				Provider: "github",
				Code:     "",
				State:    "eyJvcmdJRCI6IjAxSE...",
			},
			wantErr: true,
			errMsg:  "code is required",
		},
		{
			name: "empty state",
			request: OAuthCallbackRequest{
				Provider: "github",
				Code:     "4/0AQlEz8xY...",
				State:    "",
			},
			wantErr: true,
			errMsg:  "state is required",
		},
		{
			name: "whitespace only provider",
			request: OAuthCallbackRequest{
				Provider: "   ",
				Code:     "4/0AQlEz8xY...",
				State:    "eyJvcmdJRCI6IjAxSE...",
			},
			wantErr: false,
		},
		{
			name: "whitespace only code",
			request: OAuthCallbackRequest{
				Provider: "github",
				Code:     "   ",
				State:    "eyJvcmdJRCI6IjAxSE...",
			},
			wantErr: true,
			errMsg:  "code is required",
		},
		{
			name: "whitespace only state",
			request: OAuthCallbackRequest{
				Provider: "github",
				Code:     "4/0AQlEz8xY...",
				State:    "   ",
			},
			wantErr: true,
			errMsg:  "state is required",
		},
		{
			name: "case insensitive provider",
			request: OAuthCallbackRequest{
				Provider: "SLACK",
				Code:     "xoxp-1234567890",
				State:    "eyJvcmdJRCI6IjAxSE...",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				// Check that provider is normalized to lowercase
				assert.Equal(t, tt.request.Provider, strings.ToLower(tt.request.Provider))
			}
		})
	}
}

func TestIntegrationToken_IsExpired(t *testing.T) {
	tests := []struct {
		name     string
		token    IntegrationToken
		expected bool
	}{
		{
			name: "not expired - future expiry",
			token: IntegrationToken{
				Provider:    "github",
				AccessToken: "token123",
				ExpiresAt:   timePtr(timeInFuture(3600)), // 1 hour from now
			},
			expected: false,
		},
		{
			name: "expired - past expiry",
			token: IntegrationToken{
				Provider:    "github",
				AccessToken: "token123",
				ExpiresAt:   timePtr(timeInPast(3600)), // 1 hour ago
			},
			expected: true,
		},
		{
			name: "no expiry - never expires",
			token: IntegrationToken{
				Provider:    "github",
				AccessToken: "token123",
				ExpiresAt:   nil,
			},
			expected: false,
		},
		{
			name: "empty token - should not matter for expiry check",
			token: IntegrationToken{
				Provider:    "github",
				AccessToken: "",
				ExpiresAt:   timePtr(timeInFuture(3600)),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.token.IsExpired()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIntegrationToken_HasValidToken(t *testing.T) {
	tests := []struct {
		name     string
		token    IntegrationToken
		expected bool
	}{
		{
			name: "valid token - not expired",
			token: IntegrationToken{
				Provider:    "github",
				AccessToken: "token123",
				ExpiresAt:   timePtr(timeInFuture(3600)),
			},
			expected: true,
		},
		{
			name: "invalid token - expired",
			token: IntegrationToken{
				Provider:    "github",
				AccessToken: "token123",
				ExpiresAt:   timePtr(timeInPast(3600)),
			},
			expected: false,
		},
		{
			name: "invalid token - empty access token",
			token: IntegrationToken{
				Provider:    "github",
				AccessToken: "",
				ExpiresAt:   timePtr(timeInFuture(3600)),
			},
			expected: false,
		},
		{
			name: "valid token - no expiry",
			token: IntegrationToken{
				Provider:    "github",
				AccessToken: "token123",
				ExpiresAt:   nil,
			},
			expected: true,
		},
		{
			name: "invalid token - empty provider",
			token: IntegrationToken{
				Provider:    "",
				AccessToken: "token123",
				ExpiresAt:   timePtr(timeInFuture(3600)),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.token.HasValidToken()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper functions

func timePtr(t time.Time) *time.Time {
	return &t
}

func timeInFuture(seconds int) time.Time {
	return time.Now().Add(time.Duration(seconds) * time.Second)
}

func timeInPast(seconds int) time.Time {
	return time.Now().Add(-time.Duration(seconds) * time.Second)
}

// Test validation helper functions
func TestValidationHelpers(t *testing.T) {
	t.Run("missing required field error", func(t *testing.T) {
		err := rout.NewMissingRequiredFieldError("provider")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider")
	})

	t.Run("invalid field error", func(t *testing.T) {
		err := rout.InvalidField("redirectURI")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "redirectURI")
	})
}
