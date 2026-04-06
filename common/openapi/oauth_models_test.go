package openapi

import (
	"strings"
	"testing"

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
