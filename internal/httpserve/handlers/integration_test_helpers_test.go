package handlers_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/integrations/config"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/internal/integrations/activation"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/keymaker"
	"github.com/theopenlane/core/internal/keystore"
)

func (suite *HandlerTestSuite) withIntegrationRegistry(t *testing.T, specs map[types.ProviderType]config.ProviderSpec) func() {
	t.Helper()

	originalRegistry := suite.h.IntegrationRegistry
	originalActivation := suite.h.IntegrationActivation

	ctx := context.Background()
	reg, err := registry.NewRegistry(ctx, nil)
	require.NoError(t, err)

	for provider, spec := range specs {
		pt := provider
		builder := providers.BuilderFunc{
			ProviderType: pt,
			BuildFunc: func(context.Context, config.ProviderSpec) (providers.Provider, error) {
				return &testProvider{providerType: pt}, nil
			},
		}
		require.NoError(t, reg.UpsertProvider(ctx, spec, builder))
	}

	suite.h.IntegrationRegistry = reg

	store := keystore.NewStore(suite.db)
	sessions := keymaker.NewMemorySessionStore()
	svc, err := keymaker.NewService(reg, store, sessions, keymaker.ServiceOptions{})
	require.NoError(t, err)

	mockOps := &mockOperationRunner{}
	mockMinter := &mockPayloadMinter{}
	activationSvc, err := activation.NewService(svc, store, mockOps, mockMinter)
	require.NoError(t, err)
	suite.h.IntegrationActivation = activationSvc

	return func() {
		suite.h.IntegrationRegistry = originalRegistry
		suite.h.IntegrationActivation = originalActivation
	}
}

// mockOperationRunner implements activation.OperationRunner for tests
type mockOperationRunner struct{}

func (m *mockOperationRunner) Run(_ context.Context, _ types.OperationRequest) (types.OperationResult, error) {
	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: "mock health check passed",
		Details: map[string]any{"mock": true},
	}, nil
}

func (m *mockOperationRunner) RunWithPayload(_ context.Context, _ types.OperationRequest, _ types.CredentialPayload) (types.OperationResult, error) {
	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: "mock health check passed",
		Details: map[string]any{"mock": true},
	}, nil
}

// mockPayloadMinter implements activation.PayloadMinter for tests
type mockPayloadMinter struct{}

func (m *mockPayloadMinter) MintPayload(_ context.Context, subject types.CredentialSubject) (types.CredentialPayload, error) {
	return subject.Credential, nil
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
