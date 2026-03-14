package buildkite

import (
	"context"
	"encoding/json"
	"fmt"

	buildkitego "github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrationsv2/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const sampleSize = 10

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

// buildBuildkiteClient builds the Buildkite REST API client for one installation
func buildBuildkiteClient(_ context.Context, req types.ClientBuildRequest) (any, error) {
	var cred credential
	if err := jsonx.UnmarshalIfPresent(req.Credential.ProviderData, &cred); err != nil {
		return nil, err
	}

	if cred.APIToken == "" {
		return nil, ErrAPITokenMissing
	}

	client, err := buildkitego.NewOpts(buildkitego.WithTokenAuth(cred.APIToken))
	if err != nil {
		return nil, err
	}

	return client, nil
}

// runHealthOperation validates the Buildkite API token by calling the /v2/user endpoint
func runHealthOperation(_ context.Context, _ *generated.Integration, _ types.CredentialSet, client any, _ json.RawMessage) (json.RawMessage, error) {
	bkClient, ok := client.(*buildkitego.Client)
	if !ok {
		return nil, ErrClientType
	}

	u, _, err := bkClient.User.Get()
	if err != nil {
		return nil, fmt.Errorf("buildkite: user lookup failed: %w", err)
	}

	name := lo.FromPtrOr(u.Name, "")
	email := lo.FromPtrOr(u.Email, "")
	id := lo.FromPtrOr(u.ID, "")

	return jsonx.ToRawMessage(buildkiteHealthDetails{
		ID:    id,
		Name:  name,
		Email: email,
	})
}

// runOrganizationsCollectOperation collects Buildkite organizations for reporting
func runOrganizationsCollectOperation(_ context.Context, _ *generated.Integration, _ types.CredentialSet, client any, _ json.RawMessage) (json.RawMessage, error) {
	bkClient, ok := client.(*buildkitego.Client)
	if !ok {
		return nil, ErrClientType
	}

	orgs, _, err := bkClient.Organizations.List(nil)
	if err != nil {
		return nil, fmt.Errorf("buildkite: organizations fetch failed: %w", err)
	}

	samples := lo.Map(orgs[:min(len(orgs), sampleSize)], func(org buildkitego.Organization, _ int) buildkiteOrganizationSample {
		return buildkiteOrganizationSample{
			ID:   lo.FromPtrOr(org.ID, ""),
			Name: lo.FromPtrOr(org.Name, ""),
			Slug: lo.FromPtrOr(org.Slug, ""),
			URL:  lo.FromPtrOr(org.WebURL, ""),
		}
	})

	return jsonx.ToRawMessage(buildkiteOrganizationsDetails{
		Count:   len(orgs),
		Samples: samples,
	})
}
