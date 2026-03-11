package buildkite

import (
	"context"
	"encoding/json"
	"fmt"

	buildkitego "github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/apikey"
	"github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	// ClientBuildkiteAPI identifies the Buildkite HTTP API client
	ClientBuildkiteAPI types.ClientName = "api"

	// TypeBuildkite identifies the Buildkite provider
	TypeBuildkite types.ProviderType = "buildkite"

	buildkiteOperationHealth types.OperationName = types.OperationHealthDefault
	buildkiteOperationOrgs   types.OperationName = "organizations.collect"
)

type buildkiteHealthDetails struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type buildkiteOrganizationSample struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	URL  string `json:"url"`
}

type buildkiteOrganizationsDetails struct {
	Count   int                           `json:"count"`
	Samples []buildkiteOrganizationSample `json:"samples"`
}

// buildkiteCredentialsSchema is the JSON Schema for Buildkite organization credentials.
var buildkiteCredentialsSchema = []byte(`{"type":"object","additionalProperties":false,"required":["apiToken","organizationSlug"],"properties":{"alias":{"type":"string","title":"Credential Alias","description":"Friendly identifier for this Buildkite organization."},"apiToken":{"type":"string","title":"API Token","description":"Buildkite API token with read access to pipelines, teams, and build metadata."},"organizationSlug":{"type":"string","title":"Organization Slug","description":"Buildkite organization slug (e.g. openlane)."},"teamSlug":{"type":"string","title":"Team Slug","description":"Optional team slug used to scope which pipelines are synchronized."},"pipelineSlug":{"type":"string","title":"Pipeline Slug","description":"Optional pipeline slug when limiting sync to a single pipeline."},"syncInterval":{"type":"string","title":"Sync Interval","description":"Optional duration string indicating how frequently Buildkite metadata should be refreshed."}}}`)

// Builder returns the providers.Builder for the Buildkite provider
func Builder() providers.Builder {
	return providers.BuilderFunc{
		ProviderType: TypeBuildkite,
		SpecFunc:     buildkiteSpec,
		BuildFunc: func(ctx context.Context, s spec.ProviderSpec) (types.Provider, error) {
			return apikey.Builder(
				TypeBuildkite,
				apikey.WithTokenField("apiToken"),
				apikey.WithClientDescriptors(buildkiteClientDescriptors()),
				apikey.WithOperations(buildkiteOperations()),
			).Build(ctx, s)
		},
	}
}

// buildkiteSpec returns the static provider specification for the Buildkite provider.
func buildkiteSpec() spec.ProviderSpec {
	return spec.ProviderSpec{
		Name:        "buildkite",
		DisplayName: "Buildkite",
		Category:    "ci",
		AuthType:    types.AuthKindAPIKey,
		Active:      lo.ToPtr(false),
		Visible:     lo.ToPtr(true),
		LogoURL:     "",
		DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/buildkite/overview",
		Labels: map[string]string{
			"vendor":  "buildkite",
			"product": "pipelines",
		},
		CredentialsSchema: buildkiteCredentialsSchema,
		Description:       "Collect Buildkite organization and pipeline context to support CI security and compliance posture reporting.",
	}
}

func buildkiteClientDescriptors() []types.ClientDescriptor {
	return providerkit.DefaultClientDescriptors(TypeBuildkite, ClientBuildkiteAPI, "Buildkite REST API client", buildBuildkiteClient)
}

func buildBuildkiteClient(_ context.Context, payload types.CredentialSet, _ json.RawMessage) (types.ClientInstance, error) {
	token, err := providerkit.APITokenFromCredential(payload)
	if err != nil {
		return types.EmptyClientInstance(), err
	}

	client, err := buildkitego.NewOpts(buildkitego.WithTokenAuth(token))
	if err != nil {
		return types.EmptyClientInstance(), err
	}

	return types.NewClientInstance(client), nil
}

func buildkiteOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		providerkit.HealthOperation(buildkiteOperationHealth, "Validate Buildkite token by calling the /v2/user endpoint.", ClientBuildkiteAPI, runBuildkiteHealth),
		{
			Name:        buildkiteOperationOrgs,
			Kind:        types.OperationKindCollectFindings,
			Description: "Collect Buildkite organizations for reporting.",
			Client:      ClientBuildkiteAPI,
			Run:         runBuildkiteOrganizationsOperation,
		},
	}
}

func resolveBuildkiteClient(input types.OperationInput) (*buildkitego.Client, error) {
	if c, ok := types.ClientInstanceAs[*buildkitego.Client](input.Client); ok {
		return c, nil
	}

	token, err := providerkit.APITokenFromCredential(input.Credential)
	if err != nil {
		return nil, err
	}

	client, err := buildkitego.NewOpts(buildkitego.WithTokenAuth(token))
	if err != nil {
		return nil, err
	}

	return client, nil
}

func runBuildkiteHealth(_ context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, err := resolveBuildkiteClient(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	user, _, err := client.User.Get()
	if err != nil {
		return providerkit.OperationFailure("Buildkite user lookup failed", err, nil)
	}

	name := lo.FromPtrOr(user.Name, "")
	email := lo.FromPtrOr(user.Email, "")
	id := lo.FromPtrOr(user.ID, "")

	return providerkit.OperationSuccess(fmt.Sprintf("Buildkite token valid for %s", name), buildkiteHealthDetails{
		ID:    id,
		Name:  name,
		Email: email,
	}), nil
}

func runBuildkiteOrganizationsOperation(_ context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, err := resolveBuildkiteClient(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	orgs, _, err := client.Organizations.List(nil)
	if err != nil {
		return providerkit.OperationFailure("Buildkite organizations fetch failed", err, nil)
	}

	samples := lo.Map(orgs[:min(len(orgs), providerkit.DefaultSampleSize)], func(org buildkitego.Organization, _ int) buildkiteOrganizationSample {
		return buildkiteOrganizationSample{
			ID:   lo.FromPtrOr(org.ID, ""),
			Name: lo.FromPtrOr(org.Name, ""),
			Slug: lo.FromPtrOr(org.Slug, ""),
			URL:  lo.FromPtrOr(org.WebURL, ""),
		}
	})

	return providerkit.OperationSuccess(fmt.Sprintf("Discovered %d Buildkite organizations", len(orgs)), buildkiteOrganizationsDetails{
		Count:   len(orgs),
		Samples: samples,
	}), nil
}
