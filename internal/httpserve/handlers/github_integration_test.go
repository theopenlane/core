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

func TestValidateGithubIntegrationToken(t *testing.T) {
	tests := []struct {
		name           string
		githubResponse string
		emailsResponse string
		expectedUser   *IntegrationUserInfo
		expectError    bool
	}{
		{
			name: "successful parsing with public email",
			githubResponse: `{
				"id": 12345,
				"login": "testuser",
				"email": "testuser@example.com",
				"name": "Test User",
				"company": "Test Company"
			}`,
			expectedUser: &IntegrationUserInfo{
				ID:       "12345",
				Username: "testuser",
				Email:    "testuser@example.com",
			},
			expectError: false,
		},
		{
			name: "successful parsing with private email",
			githubResponse: `{
				"id": 67890,
				"login": "privateuser",
				"email": null,
				"name": "Private User"
			}`,
			emailsResponse: `[
				{
					"email": "privateuser@example.com",
					"primary": true,
					"verified": true
				}
			]`,
			expectedUser: &IntegrationUserInfo{
				ID:       "67890",
				Username: "privateuser",
				Email:    "privateuser@example.com",
			},
			expectError: false,
		},
		{
			name: "successful parsing without email",
			githubResponse: `{
				"id": 11111,
				"login": "noemailuser",
				"email": null,
				"name": "No Email User"
			}`,
			emailsResponse: `[]`,
			expectedUser: &IntegrationUserInfo{
				ID:       "11111",
				Username: "noemailuser",
				Email:    "",
			},
			expectError: false,
		},
		{
			name: "invalid JSON response",
			githubResponse: `{
				"id": "invalid",
				"login": 12345
			}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock GitHub API server
			githubServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/user":
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(tt.githubResponse))
				case "/user/emails":
					if tt.emailsResponse != "" {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(tt.emailsResponse))
					} else {
						w.WriteHeader(http.StatusNotFound)
					}
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer githubServer.Close()

			// Create handler and test the validation
			handler := &Handler{}
			token := &oauth2.Token{AccessToken: "test-token"}
			result, err := handler.validateGithubIntegrationTokenWithURL(context.Background(), token, githubServer.URL)

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

func TestGitHubEmailParsing(t *testing.T) {
	tests := []struct {
		name           string
		emailsResponse string
		expectedEmail  string
		expectError    bool
	}{
		{
			name: "primary verified email",
			emailsResponse: `[
				{
					"email": "secondary@example.com",
					"primary": false,
					"verified": true
				},
				{
					"email": "primary@example.com",
					"primary": true,
					"verified": true
				}
			]`,
			expectedEmail: "primary@example.com",
			expectError:   false,
		},
		{
			name: "only verified email",
			emailsResponse: `[
				{
					"email": "unverified@example.com",
					"primary": true,
					"verified": false
				},
				{
					"email": "verified@example.com",
					"primary": false,
					"verified": true
				}
			]`,
			expectedEmail: "verified@example.com",
			expectError:   false,
		},
		{
			name: "any email as fallback",
			emailsResponse: `[
				{
					"email": "fallback@example.com",
					"primary": false,
					"verified": false
				}
			]`,
			expectedEmail: "fallback@example.com",
			expectError:   false,
		},
		{
			name:           "no emails",
			emailsResponse: `[]`,
			expectedEmail:  "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock emails API server
			emailsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/user/emails" {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(tt.emailsResponse))
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer emailsServer.Close()

			// Create handler and test email fetching
			handler := &Handler{}
			email, err := handler.getGithubUserEmailWithURL(context.Background(), "test-token", emailsServer.URL)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, email)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedEmail, email)
			}
		})
	}
}

// Helper methods for testing with custom URLs

func (h *Handler) validateGithubIntegrationTokenWithURL(ctx context.Context, token *oauth2.Token, baseURL string) (*IntegrationUserInfo, error) {
	// Create HTTP client with the OAuth token
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+token.AccessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrValidateToken
	}

	// Parse the JSON response
	var githubUser GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		return nil, err
	}

	// Convert GitHub user to IntegrationUserInfo
	userInfo := &IntegrationUserInfo{
		ID:       fmt.Sprintf("%d", githubUser.ID), // Convert to string
		Username: githubUser.Login,
		Email:    githubUser.Email,
	}

	// GitHub's primary email might not be public, so we need to fetch it separately if empty
	if userInfo.Email == "" {
		if email, err := h.getGithubUserEmailWithURL(ctx, token.AccessToken, baseURL); err == nil {
			userInfo.Email = email
		}
	}

	return userInfo, nil
}

func (h *Handler) getGithubUserEmailWithURL(ctx context.Context, accessToken, baseURL string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/user/emails", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", ErrValidateToken
	}

	var emails []GitHubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	// Find the primary verified email
	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}

	// If no primary verified email, find any verified email
	for _, email := range emails {
		if email.Verified {
			return email.Email, nil
		}
	}

	// If no verified emails, return the first one
	if len(emails) > 0 {
		return emails[0].Email, nil
	}

	return "", ErrIntegrationNotFound
}
