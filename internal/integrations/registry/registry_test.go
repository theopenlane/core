package registry_test

import (
	"context"
	"errors"
	"testing"

	"github.com/theopenlane/core/internal/integrations/config"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestNewRegistryRegistersProvidersAndSanitizesDescriptors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	providerType := types.ProviderType("acme")
	spec := config.ProviderSpec{
		Name:        string(providerType),
		DisplayName: "Acme",
		Category:    "code",
		Active:      true,
	}

	clientDescriptors := []types.ClientDescriptor{
		{
			Name: "rest",
			Build: func(context.Context, types.CredentialPayload, map[string]any) (any, error) {
				return nil, nil
			},
		},
		{
			Name: "invalid",
		},
	}

	operationDescriptors := []types.OperationDescriptor{
		{
			Name: types.OperationName("sync"),
			Run: func(context.Context, types.OperationInput) (types.OperationResult, error) {
				return types.OperationResult{Status: types.OperationStatusOK}, nil
			},
		},
		{
			Name: "",
			Run: func(context.Context, types.OperationInput) (types.OperationResult, error) {
				return types.OperationResult{}, nil
			},
		},
	}

	provider := &testProvider{
		providerType:         providerType,
		clientDescriptors:    clientDescriptors,
		operationDescriptors: operationDescriptors,
	}

	builderCalls := 0
	builder := providers.BuilderFunc{
		ProviderType: providerType,
		BuildFunc: func(_ context.Context, spec config.ProviderSpec) (providers.Provider, error) {
			builderCalls++
			if spec.Name != string(providerType) {
				t.Fatalf("unexpected spec passed to builder: %+v", spec)
			}
			return provider, nil
		},
	}

	reg, err := registry.NewRegistry(ctx, map[types.ProviderType]config.ProviderSpec{providerType: spec}, []providers.Builder{builder})
	if err != nil {
		t.Fatalf("NewRegistry error: %v", err)
	}
	if builderCalls != 1 {
		t.Fatalf("expected builder to be invoked once, got %d", builderCalls)
	}

	gotProvider, ok := reg.Provider(providerType)
	if !ok || gotProvider != provider {
		t.Fatalf("expected provider to be registered")
	}

	clients := reg.ClientDescriptors(providerType)
	if len(clients) != 1 {
		t.Fatalf("expected 1 client descriptor, got %d", len(clients))
	}
	if clients[0].Provider != providerType {
		t.Fatalf("expected provider sanitized to %s, got %s", providerType, clients[0].Provider)
	}

	clients[0].Name = "mutated"
	if reg.ClientDescriptors(providerType)[0].Name != "rest" {
		t.Fatalf("expected client descriptor copies to be returned")
	}

	ops := reg.OperationDescriptors(providerType)
	if len(ops) != 1 {
		t.Fatalf("expected 1 operation descriptor, got %d", len(ops))
	}
	if ops[0].Provider != providerType {
		t.Fatalf("expected provider sanitized on operations, got %s", ops[0].Provider)
	}

	ops[0].Name = "mutated"
	if reg.OperationDescriptors(providerType)[0].Name != types.OperationName("sync") {
		t.Fatalf("expected operation descriptors to be copied")
	}
}

func TestNewRegistryBuilderError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	providerType := types.ProviderType("acme")
	spec := config.ProviderSpec{Name: string(providerType)}

	builder := providers.BuilderFunc{
		ProviderType: providerType,
		BuildFunc: func(context.Context, config.ProviderSpec) (providers.Provider, error) {
			return nil, errors.New("boom")
		},
	}

	_, err := registry.NewRegistry(ctx, map[types.ProviderType]config.ProviderSpec{providerType: spec}, []providers.Builder{builder})
	if err == nil {
		t.Fatalf("expected builder failure to propagate")
	}
}

func TestNewRegistryRequiresSpecs(t *testing.T) {
	t.Parallel()

	_, err := registry.NewRegistry(context.Background(), nil, nil)
	if err == nil {
		t.Fatalf("expected error when no specs supplied")
	}
}

func TestNewRegistrySkipsMissingBuilder(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	providerType := types.ProviderType("missing")
	spec := config.ProviderSpec{Name: string(providerType)}

	reg, err := registry.NewRegistry(ctx, map[types.ProviderType]config.ProviderSpec{providerType: spec}, nil)
	if err != nil {
		t.Fatalf("NewRegistry error: %v", err)
	}

	if _, ok := reg.Provider(providerType); ok {
		t.Fatalf("expected provider to be absent when no builder registered")
	}
	if descriptors := reg.ClientDescriptors(providerType); descriptors != nil {
		t.Fatalf("expected no client descriptors for missing provider, got %v", descriptors)
	}
}

type testProvider struct {
	providerType         types.ProviderType
	clientDescriptors    []types.ClientDescriptor
	operationDescriptors []types.OperationDescriptor
}

func (p *testProvider) Type() types.ProviderType {
	return p.providerType
}

func (p *testProvider) Capabilities() types.ProviderCapabilities { return types.ProviderCapabilities{} }

func (p *testProvider) BeginAuth(context.Context, types.AuthContext) (types.AuthSession, error) {
	return nil, nil
}

func (p *testProvider) Mint(context.Context, types.CredentialSubject) (types.CredentialPayload, error) {
	return types.CredentialPayload{}, nil
}

func (p *testProvider) ClientDescriptors() []types.ClientDescriptor {
	return p.clientDescriptors
}

func (p *testProvider) Operations() []types.OperationDescriptor {
	return p.operationDescriptors
}
