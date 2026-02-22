package okta

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/operations"
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
		operations.HealthOperation(oktaHealthOp, "Call Okta org endpoint to verify API token.", ClientOktaAPI, runOktaHealth),
		{
			Name:        oktaPoliciesOp,
			Kind:        types.OperationKindCollectFindings,
			Description: "Collect sign-on policy metadata for posture analysis.",
			Client:      ClientOktaAPI,
			Run:         runOktaPolicies,
		},
	}
}

// runOktaHealth verifies the Okta API token by fetching the org information
func runOktaHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client := auth.AuthenticatedClientFromAny(input.Client)
	baseURL, apiToken, err := oktaCredentials(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	endpoint := strings.TrimRight(baseURL, "/") + "/api/v1/org"
	var resp map[string]any
	if err := oktaGET(ctx, client, endpoint, apiToken, &resp); err != nil {
		return operations.OperationFailure("Okta org lookup failed", err, nil)
	}

	summary := fmt.Sprintf("Okta org %s reachable", baseURL)
	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: summary,
		Details: resp,
	}, nil
}

// runOktaPolicies collects a sample of Okta sign-on policies for reporting
func runOktaPolicies(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client := auth.AuthenticatedClientFromAny(input.Client)
	baseURL, apiToken, err := oktaCredentials(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	endpoint := strings.TrimRight(baseURL, "/") + "/api/v1/policies?type=SIGN_ON"
	var resp []map[string]any
	if err := oktaGET(ctx, client, endpoint, apiToken, &resp); err != nil {
		return operations.OperationFailure("Okta policies fetch failed", err, nil)
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

// oktaCredentials extracts the Okta base URL and API token from the credential payload
func oktaCredentials(input types.OperationInput) (string, string, error) {
	data := input.Credential.Data
	baseURL, _ := data.ProviderData["orgUrl"].(string)
	apiToken := data.APIToken
	if baseURL == "" || apiToken == "" {
		return "", "", ErrCredentialsMissing
	}

	return baseURL, apiToken, nil
}

// oktaGET performs a GET request to the Okta API and decodes the JSON response
func oktaGET(ctx context.Context, client *auth.AuthenticatedClient, endpoint, apiToken string, out any) error {
	headers := map[string]string{
		"Authorization": "SSWS " + apiToken,
	}

	if err := auth.GetJSONWithClient(ctx, client, endpoint, "", headers, out); err != nil {
		if errors.Is(err, auth.ErrHTTPRequestFailed) {
			return ErrAPIRequest
		}
		return err
	}

	return nil
}
