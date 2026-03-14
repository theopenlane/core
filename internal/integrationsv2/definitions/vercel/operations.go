package vercel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrationsv2/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const vercelAPIBaseURL = "https://api.vercel.com"

type vercelUserResponse struct {
	User struct {
		ID    string `json:"uid"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"user"`
}

type vercelHealthDetails struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type vercelProjectSample struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Framework string `json:"framework"`
}

type vercelProjectsDetails struct {
	Projects []vercelProjectSample `json:"projects"`
}

// buildVercelClient builds the Vercel REST API client for one installation
func buildVercelClient(_ context.Context, req types.ClientBuildRequest) (any, error) {
	var cred credential
	if err := jsonx.UnmarshalIfPresent(req.Credential.ProviderData, &cred); err != nil {
		return nil, err
	}

	if cred.APIToken == "" {
		return nil, ErrAPITokenMissing
	}

	return providerkit.NewAuthenticatedClient(vercelAPIBaseURL, cred.APIToken, nil), nil
}

// runHealthOperation calls /v2/user to verify the Vercel token
func runHealthOperation(ctx context.Context, _ *generated.Integration, _ types.CredentialSet, client any, _ json.RawMessage) (json.RawMessage, error) {
	c, ok := client.(*providerkit.AuthenticatedClient)
	if !ok {
		return nil, ErrClientType
	}

	var resp vercelUserResponse
	if err := c.GetJSON(ctx, "/v2/user", &resp); err != nil {
		return nil, fmt.Errorf("vercel: user lookup failed: %w", err)
	}

	return jsonx.ToRawMessage(vercelHealthDetails{
		ID:    resp.User.ID,
		Name:  resp.User.Name,
		Email: resp.User.Email,
	})
}

// runProjectsSampleOperation fetches a small sample of Vercel projects
func runProjectsSampleOperation(ctx context.Context, _ *generated.Integration, _ types.CredentialSet, client any, _ json.RawMessage) (json.RawMessage, error) {
	c, ok := client.(*providerkit.AuthenticatedClient)
	if !ok {
		return nil, ErrClientType
	}

	params := url.Values{}
	params.Set("limit", "5")

	var resp struct {
		Projects []struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Framework string `json:"framework"`
		} `json:"projects"`
	}

	if err := c.GetJSONWithParams(ctx, "/v4/projects", params, &resp); err != nil {
		return nil, fmt.Errorf("vercel: projects fetch failed: %w", err)
	}

	samples := make([]vercelProjectSample, 0, len(resp.Projects))
	for _, p := range resp.Projects {
		samples = append(samples, vercelProjectSample{
			ID:        p.ID,
			Name:      p.Name,
			Framework: p.Framework,
		})
	}

	return jsonx.ToRawMessage(vercelProjectsDetails{Projects: samples})
}
