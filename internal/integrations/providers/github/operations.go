package github

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	githubOperationHealth types.OperationName = "health.default"
	githubOperationRepos  types.OperationName = "repos.collect_metadata"

	defaultPerPage = 50
	maxPerPage     = 100
	maxSampleSize  = 5
)

func githubOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		{
			Name:        githubOperationHealth,
			Kind:        types.OperationKindHealth,
			Description: "Validate GitHub OAuth token by calling the /user endpoint.",
			Run:         runGitHubHealthOperation,
		},
		{
			Name:        githubOperationRepos,
			Kind:        types.OperationKindCollectFindings,
			Description: "Collect repository metadata for the authenticated account.",
			Run:         runGitHubRepoOperation,
			ConfigSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"visibility": map[string]any{
						"type":        "string",
						"description": "Optional visibility filter (all, public, private)",
					},
					"per_page": map[string]any{
						"type":        "integer",
						"description": "Override the number of repos fetched per page (max 100).",
					},
				},
			},
		},
	}
}

type githubUserResponse struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
	Name  string `json:"name"`
}

type githubRepoResponse struct {
	Name      string    `json:"name"`
	Private   bool      `json:"private"`
	UpdatedAt time.Time `json:"updated_at"`
	HTMLURL   string    `json:"html_url"`
}

func runGitHubHealthOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	token, err := oauthTokenFromPayload(input.Credential)
	if err != nil {
		return types.OperationResult{}, err
	}

	var user githubUserResponse
	if err := fetchGitHubResource(ctx, token, "user", nil, &user); err != nil {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "GitHub user lookup failed",
			Details: map[string]any{"error": err.Error()},
		}, err
	}

	details := map[string]any{
		"login": user.Login,
		"id":    user.ID,
		"name":  user.Name,
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("GitHub token valid for %s", user.Login),
		Details: details,
	}, nil
}

func runGitHubRepoOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	token, err := oauthTokenFromPayload(input.Credential)
	if err != nil {
		return types.OperationResult{}, err
	}

	params := url.Values{}
	params.Set("per_page", fmt.Sprintf("%d", clampPerPage(intFromConfig(input.Config, "per_page", defaultPerPage))))
	if visibility := stringFromConfig(input.Config, "visibility"); visibility != "" {
		params.Set("visibility", visibility)
	}

	var repos []githubRepoResponse
	if err := fetchGitHubResource(ctx, token, "user/repos", params, &repos); err != nil {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "GitHub repository collection failed",
			Details: map[string]any{"error": err.Error()},
		}, err
	}

	samples := make([]map[string]any, 0, minInt(maxSampleSize, len(repos)))
	for _, repo := range repos {
		if len(samples) >= cap(samples) {
			break
		}
		samples = append(samples, map[string]any{
			"name":       repo.Name,
			"private":    repo.Private,
			"updated_at": repo.UpdatedAt,
			"url":        repo.HTMLURL,
		})
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Collected %d repositories", len(repos)),
		Details: map[string]any{
			"count":   len(repos),
			"samples": samples,
			"params":  params.Encode(),
		},
	}, nil
}

func fetchGitHubResource(ctx context.Context, token, path string, params url.Values, out any) error {
	endpoint := "https://api.github.com/" + path
	if params != nil {
		if encoded := params.Encode(); encoded != "" {
			endpoint += "?" + encoded
		}
	}

	headers := map[string]string{
		"Accept": "application/vnd.github+json",
	}

	if err := helpers.HTTPGetJSON(ctx, nil, endpoint, token, headers, out); err != nil {
		if errors.Is(err, helpers.ErrHTTPRequestFailed) {
			return fmt.Errorf("%w (path %s): %s", ErrAPIRequest, path, err.Error())
		}
		return err
	}

	return nil
}

func oauthTokenFromPayload(payload types.CredentialPayload) (string, error) {
	tokenOpt := payload.OAuthTokenOption()
	if !tokenOpt.IsPresent() {
		return "", ErrOAuthTokenMissing
	}

	token := tokenOpt.MustGet()
	if token == nil || token.AccessToken == "" {
		return "", ErrAccessTokenEmpty
	}

	return token.AccessToken, nil
}

func stringFromConfig(config map[string]any, key string) string {
	if len(config) == 0 {
		return ""
	}
	if value, ok := config[key]; ok {
		if str, ok := value.(string); ok {
			return strings.TrimSpace(str)
		}
	}
	return ""
}

func intFromConfig(config map[string]any, key string, fallback int) int {
	if len(config) == 0 {
		return fallback
	}
	if value, ok := config[key]; ok {
		switch v := value.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return fallback
}

func clampPerPage(value int) int {
	if value <= 0 {
		return defaultPerPage
	}
	if value > maxPerPage {
		return maxPerPage
	}
	return value
}

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
