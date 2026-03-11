package handlers_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/internal/integrations/providers"
	githubprovider "github.com/theopenlane/core/internal/integrations/providers/github"
	"github.com/theopenlane/core/internal/integrations/registry"
	integrationruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// decodeGitHubAppConfig decodes a github.AppConfig from a spec's ProviderConfig field.
// Returns a zero-value AppConfig if ProviderConfig is empty or decoding fails.
func decodeGitHubAppConfig(t interface{ Helper(); Fatal(...any) }, provSpec spec.ProviderSpec) githubprovider.AppConfig {
	t.Helper()
	var cfg githubprovider.AppConfig
	if err := jsonx.UnmarshalIfPresent(provSpec.ProviderConfig, &cfg); err != nil {
		t.Fatal(err)
	}
	return cfg
}

func (suite *HandlerTestSuite) withIntegrationRegistry(t *testing.T, specs map[types.ProviderType]spec.ProviderSpec) func() {
	t.Helper()

	originalRuntime := suite.h.IntegrationRuntime

	ctx := context.Background()
	reg, err := registry.NewRegistry(ctx, nil)
	assert.NoError(t, err)

	for provider, providerSpec := range specs {
		pt := provider
		ps := providerSpec
		builder := providers.BuilderFunc{
			ProviderType: pt,
			BuildFunc: func(context.Context, spec.ProviderSpec) (types.Provider, error) {
				return &testProvider{providerType: pt}, nil
			},
		}
		assert.NoError(t, reg.UpsertProvider(ctx, ps, builder))
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

func (suite *HandlerTestSuite) withGitHubAppIntegrationRuntime(t *testing.T, spec spec.ProviderSpec) func() {
	t.Helper()

	originalRuntime := suite.h.IntegrationRuntime

	ctx := context.Background()
	reg, err := registry.NewRegistry(ctx, nil)
	assert.NoError(t, err)
	appCfg := decodeGitHubAppConfig(t, spec)
	assert.NoError(t, reg.UpsertProvider(ctx, spec, githubprovider.AppBuilder(appCfg)))

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

func (p *testProvider) Mint(context.Context, types.CredentialMintRequest) (types.CredentialSet, error) {
	return types.CredentialSet{}, nil
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

func gcpSCCSpec() spec.ProviderSpec {
	schema, err := jsonx.ToRawMessage(map[string]any{
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
	})
	if err != nil {
		panic(err)
	}

	return spec.ProviderSpec{
		Name:              "gcpscc",
		DisplayName:       "Google Cloud SCC",
		Category:          "cloud",
		Description:       "Google Cloud Security Command Center integration",
		AuthType:          types.AuthKindWorkloadIdentity,
		Active:            lo.ToPtr(true),
		Visible:           lo.ToPtr(true),
		Tags:              []string{"cloud", "google"},
		CredentialsSchema: schema,
	}
}
