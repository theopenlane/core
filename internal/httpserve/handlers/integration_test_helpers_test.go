package handlers_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/internal/integrations/config"
	"github.com/theopenlane/core/internal/integrations/providers"
	githubprovider "github.com/theopenlane/core/internal/integrations/providers/github"
	"github.com/theopenlane/core/internal/integrations/registry"
	integrationruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/integrations/types"
)

func (suite *HandlerTestSuite) withIntegrationRegistry(t *testing.T, specs map[types.ProviderType]config.ProviderSpec) func() {
	t.Helper()

	originalRuntime := suite.h.IntegrationRuntime

	ctx := context.Background()
	reg, err := registry.NewRegistry(ctx, nil)
	assert.NoError(t, err)

	for provider, spec := range specs {
		pt := provider
		builder := providers.BuilderFunc{
			ProviderType: pt,
			BuildFunc: func(context.Context, config.ProviderSpec) (providers.Provider, error) {
				return &testProvider{providerType: pt}, nil
			},
		}
		assert.NoError(t, reg.UpsertProvider(ctx, spec, builder))
	}

	rt, err := integrationruntime.New(integrationruntime.Config{
		Registry: reg,
		DB:       suite.db,
	})
	assert.NoError(t, err)
	suite.h.IntegrationRuntime = rt

	return func() {
		suite.h.IntegrationRuntime = originalRuntime
	}
}

func (suite *HandlerTestSuite) withGitHubAppIntegrationRuntime(t *testing.T, spec config.ProviderSpec) func() {
	t.Helper()

	originalRuntime := suite.h.IntegrationRuntime

	ctx := context.Background()
	reg, err := registry.NewRegistry(ctx, nil)
	assert.NoError(t, err)
	assert.NoError(t, reg.UpsertProvider(ctx, spec, githubprovider.AppBuilder()))

	rt, err := integrationruntime.New(integrationruntime.Config{
		Registry:           reg,
		DB:                 suite.db,
		SuccessRedirectURL: spec.SuccessRedirectURL,
	})
	assert.NoError(t, err)
	suite.h.IntegrationRuntime = rt

	return func() {
		suite.h.IntegrationRuntime = originalRuntime
	}
}

type testProvider struct {
	providerType types.ProviderType
}

func (p *testProvider) Type() types.ProviderType { return p.providerType }

func (p *testProvider) Capabilities() types.ProviderCapabilities { return types.ProviderCapabilities{} }

func (p *testProvider) BeginAuth(context.Context, types.AuthContext) (types.AuthSession, error) {
	return nil, nil
}

func (p *testProvider) Mint(context.Context, types.CredentialSubject) (types.CredentialPayload, error) {
	return types.CredentialPayload{}, nil
}

func (p *testProvider) Operations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		{
			Provider: p.providerType,
			Name:     types.OperationHealthDefault,
			Kind:     types.OperationKindScanSettings,
			Run: func(context.Context, types.OperationInput) (types.OperationResult, error) {
				return types.OperationResult{
					Status:  types.OperationStatusOK,
					Summary: "ok",
				}, nil
			},
		},
	}
}

func gcpSCCSpec() config.ProviderSpec {
	return config.ProviderSpec{
		Name:        "gcpscc",
		DisplayName: "Google Cloud SCC",
		Category:    "cloud",
		Description: "Google Cloud Security Command Center integration",
		AuthType:    types.AuthKindWorkloadIdentity,
		Active:      lo.ToPtr(true),
		Visible:     lo.ToPtr(true),
		Tags:        []string{"cloud", "google"},
		CredentialsSchema: map[string]any{
			"type": "object",
			"required": []string{
				"projectId",
				"serviceAccountEmail",
			},
			"properties": map[string]any{
				"projectId": map[string]any{
					"type":        "string",
					"title":       "Project ID",
					"description": "GCP project identifier",
				},
				"serviceAccountEmail": map[string]any{
					"type":        "string",
					"title":       "Service Account Email",
					"description": "Workload identity service account",
				},
				"organizationId": map[string]any{
					"type":        "string",
					"title":       "Organization ID",
					"description": "Optional organization scope",
				},
			},
		},
	}
}
