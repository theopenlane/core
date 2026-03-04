package buildkite

import (
	"context"
	"fmt"

	buildkitego "github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	buildkiteOperationHealth types.OperationName = "health.default"
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

// buildkiteOperations returns the Buildkite operations supported by this provider.
func buildkiteOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		operations.HealthOperation(buildkiteOperationHealth, "Validate Buildkite token by calling the /v2/user endpoint.", ClientBuildkiteAPI, runBuildkiteHealth),
		{
			Name:        buildkiteOperationOrgs,
			Kind:        types.OperationKindCollectFindings,
			Description: "Collect Buildkite organizations for reporting.",
			Client:      ClientBuildkiteAPI,
			Run:         runBuildkiteOrganizationsOperation,
		},
	}
}

// resolveBuildkiteClient returns a pooled Buildkite client or builds one from the credential payload.
func resolveBuildkiteClient(input types.OperationInput) (*buildkitego.Client, error) {
	if c, ok := types.ClientInstanceAs[*buildkitego.Client](input.Client); ok {
		return c, nil
	}

	token, err := auth.APITokenFromPayload(input.Credential)
	if err != nil {
		return nil, err
	}

	client, err := buildkitego.NewOpts(buildkitego.WithTokenAuth(token))
	if err != nil {
		return nil, err
	}

	return client, nil
}

// runBuildkiteHealth validates the Buildkite token by fetching the current user.
func runBuildkiteHealth(_ context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, err := resolveBuildkiteClient(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	user, _, err := client.User.Get()
	if err != nil {
		return operations.OperationFailure("Buildkite user lookup failed", err, nil)
	}

	name := lo.FromPtrOr(user.Name, "")
	email := lo.FromPtrOr(user.Email, "")
	id := lo.FromPtrOr(user.ID, "")

	return operations.OperationSuccess(fmt.Sprintf("Buildkite token valid for %s", name), buildkiteHealthDetails{
		ID:    id,
		Name:  name,
		Email: email,
	}), nil
}

// runBuildkiteOrganizationsOperation collects Buildkite org metadata for reporting.
func runBuildkiteOrganizationsOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, err := resolveBuildkiteClient(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	orgs, _, err := client.Organizations.List(nil)
	if err != nil {
		return operations.OperationFailure("Buildkite organizations fetch failed", err, nil)
	}

	samples := lo.Map(orgs[:min(len(orgs), operations.DefaultSampleSize)], func(org buildkitego.Organization, _ int) buildkiteOrganizationSample {
		return buildkiteOrganizationSample{
			ID:   lo.FromPtrOr(org.ID, ""),
			Name: lo.FromPtrOr(org.Name, ""),
			Slug: lo.FromPtrOr(org.Slug, ""),
			URL:  lo.FromPtrOr(org.WebURL, ""),
		}
	})

	return operations.OperationSuccess(fmt.Sprintf("Discovered %d Buildkite organizations", len(orgs)), buildkiteOrganizationsDetails{
		Count:   len(orgs),
		Samples: samples,
	}), nil
}
