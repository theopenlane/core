package okta

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	oktaHealthOp   types.OperationName = "health.default"
	oktaPoliciesOp types.OperationName = "policies.collect"
)
const maxSampleSize = 5

// oktaOperations returns the Okta operations supported by this provider.
func oktaOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		{
			Name:        oktaHealthOp,
			Kind:        types.OperationKindHealth,
			Description: "Call Okta org endpoint to verify API token.",
			Client:      ClientOktaAPI,
			Run:         runOktaHealth,
		},
		{
			Name:        oktaPoliciesOp,
			Kind:        types.OperationKindCollectFindings,
			Description: "Collect sign-on policy metadata for posture analysis.",
			Client:      ClientOktaAPI,
			Run:         runOktaPolicies,
		},
	}
}

func runOktaHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client := helpers.AuthenticatedClientFromAny(input.Client)
	baseURL, apiToken, err := oktaCredentials(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	endpoint := strings.TrimRight(baseURL, "/") + "/api/v1/org"
	var resp map[string]any
	if client != nil {
		if err := client.GetJSON(ctx, endpoint, &resp); err != nil {
			return types.OperationResult{
				Status:  types.OperationStatusFailed,
				Summary: "Okta org lookup failed",
				Details: map[string]any{"error": err.Error()},
			}, err
		}
	} else if err := oktaGET(ctx, endpoint, apiToken, &resp); err != nil {
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

// runOktaPolicies collects a sample of Okta sign-on policies for reporting.
func runOktaPolicies(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client := helpers.AuthenticatedClientFromAny(input.Client)
	baseURL, apiToken, err := oktaCredentials(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	endpoint := strings.TrimRight(baseURL, "/") + "/api/v1/policies?type=SIGN_ON"
	var resp []map[string]any
	if client != nil {
		if err := client.GetJSON(ctx, endpoint, &resp); err != nil {
			return types.OperationResult{
				Status:  types.OperationStatusFailed,
				Summary: "Okta policies fetch failed",
				Details: map[string]any{"error": err.Error()},
			}, err
		}
	} else if err := oktaGET(ctx, endpoint, apiToken, &resp); err != nil {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "Okta policies fetch failed",
			Details: map[string]any{"error": err.Error()},
		}, err
	}

	samples := resp
	if len(samples) > maxSampleSize {
		samples = samples[:maxSampleSize]
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
		return "", "", ErrCredentialsMissing
	}
	return baseURL, apiToken, nil
}

func oktaGET(ctx context.Context, endpoint, apiToken string, out any) error {
	headers := map[string]string{
		"Authorization": "SSWS " + apiToken,
	}

	if err := helpers.HTTPGetJSON(ctx, nil, endpoint, "", headers, out); err != nil {
		if errors.Is(err, helpers.ErrHTTPRequestFailed) {
			return fmt.Errorf("%w: %w", ErrAPIRequest, err)
		}
		return err
	}

	return nil
}
