package vercel

import (
	"context"
	"fmt"
	"net/url"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/apikey"
	"github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	// ClientVercelAPI identifies the Vercel HTTP API client
	ClientVercelAPI types.ClientName = "api"

	// TypeVercel identifies the Vercel provider
	TypeVercel types.ProviderType = "vercel"

	vercelHealthOp   types.OperationName = types.OperationHealthDefault
	vercelProjectsOp types.OperationName = "projects.sample"
	vercelAPIBaseURL                     = "https://api.vercel.com"
)

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

// vercelCredentialsSchema is the JSON Schema for Vercel project credentials.
var vercelCredentialsSchema = []byte(`{"type":"object","additionalProperties":false,"required":["apiToken"],"properties":{"alias":{"type":"string","title":"Credential Alias","description":"Friendly name shown inside Openlane for this token."},"apiToken":{"type":"string","title":"API Token","description":"Vercel personal or team API token with read access to projects and settings."},"teamId":{"type":"string","title":"Team ID","description":"Optional team identifier used if the token is scoped to a team."},"projectId":{"type":"string","title":"Project ID","description":"Optional project identifier to scope compliance checks."},"environment":{"type":"string","title":"Default Environment","description":"Indicates which deployment environment should be queried first.","enum":["production","preview","development"],"default":"production"},"captureDeployHooks":{"type":"boolean","title":"Capture Deploy Hooks","description":"When enabled, deploy hook metadata is persisted alongside the integration.","default":false},"metadataTags":{"type":"array","title":"Metadata Tags","description":"Optional list of tags that categorize the projects pulled from Vercel.","items":{"type":"string"}}}}`)

// Builder returns the providers.Builder for the Vercel provider
func Builder() providers.Builder {
	return providers.BuilderFunc{
		ProviderType: TypeVercel,
		SpecFunc:     vercelSpec,
		BuildFunc: func(ctx context.Context, s spec.ProviderSpec) (types.Provider, error) {
			return apikey.Builder(
				TypeVercel,
				apikey.WithTokenField("apiToken"),
				apikey.WithClientDescriptors(vercelClientDescriptors()),
				apikey.WithOperations(vercelOperations()),
			).Build(ctx, s)
		},
	}
}

// vercelSpec returns the static provider specification for the Vercel provider.
func vercelSpec() spec.ProviderSpec {
	return spec.ProviderSpec{
		Name:        "vercel",
		DisplayName: "Vercel",
		Category:    "devops",
		AuthType:    types.AuthKindAPIKey,
		Active:      lo.ToPtr(false),
		Visible:     lo.ToPtr(true),
		LogoURL:     "",
		DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/vercel/overview",
		Labels: map[string]string{
			"vendor":  "vercel",
			"product": "deployment",
		},
		CredentialsSchema: vercelCredentialsSchema,
		Description:       "Collect Vercel project and deployment context to support devops posture and drift detection workflows.",
	}
}

func vercelClientDescriptors() []types.ClientDescriptor {
	return providerkit.DefaultClientDescriptors(TypeVercel, ClientVercelAPI, "Vercel REST API client", providerkit.TokenClientBuilder(providerkit.APITokenFromCredential, nil))
}

func vercelOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		providerkit.HealthOperation(vercelHealthOp, "Call Vercel /v2/user to verify token and account.", ClientVercelAPI,
			providerkit.HealthCheckRunner(providerkit.APITokenFromCredential, "https://api.vercel.com/v2/user", "Vercel user lookup failed",
				func(resp vercelUserResponse) (string, any) {
					return fmt.Sprintf("Vercel token valid for %s", resp.User.Email), vercelHealthDetails{
						ID:    resp.User.ID,
						Name:  resp.User.Name,
						Email: resp.User.Email,
					}
				})),
		{
			Name:        vercelProjectsOp,
			Kind:        types.OperationKindCollectFindings,
			Description: "Collect a sample of Vercel projects for drift detection.",
			Client:      ClientVercelAPI,
			Run:         runVercelProjects,
		},
	}
}

func runVercelProjects(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, err := providerkit.ResolveAuthenticatedClient(input, providerkit.APITokenFromCredential, vercelAPIBaseURL, nil)
	if err != nil {
		return types.OperationResult{}, err
	}

	params := url.Values{}
	params.Set("limit", "5")

	var resp struct {
		// Projects lists projects returned by the API
		Projects []struct {
			// ID is the project identifier
			ID string `json:"id"`
			// Name is the project name
			Name string `json:"name"`
			// Framework is the detected framework name
			Framework string `json:"framework"`
		} `json:"projects"`
	}

	if err := client.GetJSONWithParams(ctx, "/v4/projects", params, &resp); err != nil {
		return providerkit.OperationFailure("Vercel projects fetch failed", err, nil)
	}

	samples := make([]vercelProjectSample, 0, len(resp.Projects))
	for _, project := range resp.Projects {
		samples = append(samples, vercelProjectSample{
			ID:        project.ID,
			Name:      project.Name,
			Framework: project.Framework,
		})
	}

	return providerkit.OperationSuccess(fmt.Sprintf("Fetched %d Vercel projects", len(samples)), vercelProjectsDetails{Projects: samples}), nil
}
