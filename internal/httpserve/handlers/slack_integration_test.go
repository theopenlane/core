package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestValidateSlackIntegrationToken(t *testing.T) {
	tests := []struct {
		name          string
		slackResponse string
		expectedUser  *IntegrationUserInfo
		expectError   bool
	}{
		{
			name: "successful parsing with email",
			slackResponse: `{
				"ok": true,
				"user": {
					"id": "U12345",
					"team_id": "T12345",
					"name": "testuser",
					"deleted": false,
					"real_name": "Test User",
					"profile": {
						"email": "testuser@example.com",
						"display_name": "Test User",
						"display_name_normalized": "test user",
						"real_name": "Test User",
						"real_name_normalized": "test user"
					}
				}
			}`,
			expectedUser: &IntegrationUserInfo{
				ID:       "U12345",
				Username: "testuser",
				Email:    "testuser@example.com",
			},
			expectError: false,
		},
		{
			name: "successful parsing with display name fallback",
			slackResponse: `{
				"ok": true,
				"user": {
					"id": "U67890",
					"team_id": "T67890",
					"name": "",
					"deleted": false,
					"real_name": "Another User",
					"profile": {
						"email": "another@example.com",
						"display_name": "AnotherUser",
						"display_name_normalized": "anotheruser",
						"real_name": "Another User",
						"real_name_normalized": "another user"
					}
				}
			}`,
			expectedUser: &IntegrationUserInfo{
				ID:       "U67890",
				Username: "AnotherUser",
				Email:    "another@example.com",
			},
			expectError: false,
		},
		{
			name: "error response from Slack API",
			slackResponse: `{
				"ok": false,
				"error": "invalid_auth"
			}`,
			expectError: true,
		},
		{
			name: "invalid JSON response",
			slackResponse: `{
				"ok": true,
				"user": {
					"id": "not a valid structure
				}
			}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock Slack API server
			slackServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/users.identity", r.URL.Path)
				assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.slackResponse))
			}))
			defer slackServer.Close()

			// Create handler and test the validation
			handler := &Handler{}
			token := &oauth2.Token{AccessToken: "test-token"}
			result, err := handler.validateSlackIntegrationTokenWithURL(context.Background(), token, slackServer.URL)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedUser.ID, result.ID)
				assert.Equal(t, tt.expectedUser.Username, result.Username)
				assert.Equal(t, tt.expectedUser.Email, result.Email)
			}
		})
	}
}

// Helper method for testing with custom URL
func (h *Handler) validateSlackIntegrationTokenWithURL(ctx context.Context, token *oauth2.Token, baseURL string) (*IntegrationUserInfo, error) {
	// Create HTTP client with the OAuth token
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/users.identity", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Slack API request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Slack API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Slack API returned status %d: %w", resp.StatusCode, ErrValidateToken)
	}

	// Parse the JSON response
	var slackResp SlackUser
	if err := json.NewDecoder(resp.Body).Decode(&slackResp); err != nil {
		return nil, fmt.Errorf("failed to parse Slack API response: %w", err)
	}

	// Check if the response is successful
	if !slackResp.OK {
		return nil, fmt.Errorf("Slack API error: %s", slackResp.Error)
	}

	// Convert Slack user to IntegrationUserInfo
	userInfo := &IntegrationUserInfo{
		ID:       slackResp.User.ID,
		Username: slackResp.User.Name,
		Email:    slackResp.User.Profile.Email,
	}

	// Use display name if username is empty
	if userInfo.Username == "" && slackResp.User.Profile.DisplayName != "" {
		userInfo.Username = slackResp.User.Profile.DisplayName
	}

	return userInfo, nil
}
