package okta

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	oktaHealthOp   types.OperationName = "health.default"
	oktaPoliciesOp types.OperationName = "policies.collect"
)

var oktaHTTPClient = &http.Client{Timeout: 10 * time.Second}

func oktaOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		{
			Name:        oktaHealthOp,
			Kind:        types.OperationKindHealth,
			Description: "Call Okta org endpoint to verify API token.",
			Run:         runOktaHealth,
		},
		{
			Name:        oktaPoliciesOp,
			Kind:        types.OperationKindCollectFindings,
			Description: "Collect sign-on policy metadata for posture analysis.",
			Run:         runOktaPolicies,
		},
	}
}

func runOktaHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	baseURL, apiToken, err := oktaCredentials(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	endpoint := strings.TrimRight(baseURL, "/") + "/api/v1/org"
	var resp map[string]any
	if err := oktaGET(ctx, endpoint, apiToken, &resp); err != nil {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "Okta org lookup failed",
			Details: map[string]any{"error": err.Error()},
		}, err
	}

	summary := fmt.Sprintf("Okta org %s reachable", baseURL)
	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: summary,
		Details: resp,
	}, nil
}

func runOktaPolicies(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	baseURL, apiToken, err := oktaCredentials(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	endpoint := strings.TrimRight(baseURL, "/") + "/api/v1/policies?type=SIGN_ON"
	var resp []map[string]any
	if err := oktaGET(ctx, endpoint, apiToken, &resp); err != nil {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "Okta policies fetch failed",
			Details: map[string]any{"error": err.Error()},
		}, err
	}

	samples := resp
	if len(samples) > 5 {
		samples = samples[:5]
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Collected %d sign-on policies", len(resp)),
		Details: map[string]any{
			"count":   len(resp),
			"samples": samples,
		},
	}, nil
}

func oktaCredentials(input types.OperationInput) (string, string, error) {
	data := input.Credential.Data
	baseURL := ""
	if data.ProviderData != nil {
		if value, ok := data.ProviderData["orgUrl"].(string); ok {
			baseURL = strings.TrimSpace(value)
		}
	}
	apiToken := strings.TrimSpace(data.APIToken)
	if baseURL == "" || apiToken == "" {
		return "", "", fmt.Errorf("okta: org url or api token missing")
	}
	return baseURL, apiToken, nil
}

func oktaGET(ctx context.Context, endpoint, apiToken string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "SSWS "+apiToken)
	req.Header.Set("Accept", "application/json")

	resp, err := oktaHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("okta api %s: %s", endpoint, resp.Status)
	}

	if out == nil {
		return nil
	}

	dec := json.NewDecoder(resp.Body)
	return dec.Decode(out)
}
