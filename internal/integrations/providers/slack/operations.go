package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	slackOperationHealth types.OperationName = "health.default"
	slackOperationTeam   types.OperationName = "team.inspect"
)

var slackHTTPClient = &http.Client{Timeout: 10 * time.Second}

func slackOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		{
			Name:        slackOperationHealth,
			Kind:        types.OperationKindHealth,
			Description: "Call auth.test to ensure the Slack token is valid and scoped correctly.",
			Run:         runSlackHealthOperation,
		},
		{
			Name:        slackOperationTeam,
			Kind:        types.OperationKindScanSettings,
			Description: "Collect workspace metadata via team.info for posture analysis.",
			Run:         runSlackTeamOperation,
		},
	}
}

type slackAuthTestResponse struct {
	OK    bool   `json:"ok"`
	URL   string `json:"url"`
	Team  string `json:"team"`
	User  string `json:"user"`
	Error string `json:"error"`
}

type slackTeamInfoResponse struct {
	OK    bool          `json:"ok"`
	Team  slackTeamInfo `json:"team"`
	Error string        `json:"error"`
}

type slackTeamInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	EmailDomain string `json:"email_domain"`
}

func runSlackHealthOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	token, err := oauthTokenFromPayload(input.Credential)
	if err != nil {
		return types.OperationResult{}, err
	}

	var resp slackAuthTestResponse
	if err := slackAPIGet(ctx, token, "auth.test", nil, &resp); err != nil {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "Slack auth.test failed",
			Details: map[string]any{"error": err.Error()},
		}, err
	}

	if !resp.OK {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "Slack auth.test returned error",
			Details: map[string]any{"error": resp.Error},
		}, fmt.Errorf("%w: %s", ErrSlackAPIError, resp.Error)
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Slack token valid for workspace %s", resp.Team),
		Details: map[string]any{
			"team": resp.Team,
			"url":  resp.URL,
			"user": resp.User,
		},
	}, nil
}

func runSlackTeamOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	token, err := oauthTokenFromPayload(input.Credential)
	if err != nil {
		return types.OperationResult{}, err
	}

	var resp slackTeamInfoResponse
	if err := slackAPIGet(ctx, token, "team.info", nil, &resp); err != nil {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "Slack team.info failed",
			Details: map[string]any{"error": err.Error()},
		}, err
	}

	if !resp.OK {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "Slack team.info returned error",
			Details: map[string]any{"error": resp.Error},
		}, fmt.Errorf("%w: %s", ErrSlackAPIError, resp.Error)
	}

	team := resp.Team
	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Workspace %s (%s) settings retrieved", team.Name, team.ID),
		Details: map[string]any{
			"teamId":      team.ID,
			"name":        team.Name,
			"domain":      team.Domain,
			"emailDomain": team.EmailDomain,
		},
	}, nil
}

func slackAPIGet(ctx context.Context, token, method string, params url.Values, out any) error {
	endpoint := "https://slack.com/api/" + method
	if params != nil {
		if query := params.Encode(); query != "" {
			endpoint += "?" + query
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := slackHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("%w (method %s): %s", ErrAPIRequest, method, resp.Status)
	}

	if out == nil {
		return nil
	}

	dec := json.NewDecoder(resp.Body)
	return dec.Decode(out)
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
