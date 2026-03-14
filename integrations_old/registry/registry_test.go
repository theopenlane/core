package registry_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/providers/catalog"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestNewRegistry(t *testing.T) {
	ctx := context.Background()

	r, err := registry.NewRegistry(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, r)
}

func TestRegistry_Provider(t *testing.T) {
	ctx := context.Background()

	r, err := registry.NewRegistry(ctx, catalog.Builders(catalog.Config{}))
	require.NoError(t, err)

	_, ok := r.Provider(types.ProviderType("github"))
	assert.True(t, ok)

	_, ok = r.Provider(types.ProviderType("nonexistent"))
	assert.False(t, ok)
}

func TestRegistry_Config(t *testing.T) {
	ctx := context.Background()

	r, err := registry.NewRegistry(ctx, catalog.Builders(catalog.Config{}))
	require.NoError(t, err)

	providerSpec, ok := r.Config(types.ProviderType("slack"))
	assert.True(t, ok)
	assert.Equal(t, "slack", providerSpec.Name)
}

func TestRegistry_ProviderMetadataCatalog(t *testing.T) {
	ctx := context.Background()

	r, err := registry.NewRegistry(ctx, catalog.Builders(catalog.Config{}))
	require.NoError(t, err)

	providerCatalog := r.ProviderMetadataCatalog()
	assert.NotEmpty(t, providerCatalog)
}

func TestRegistry_MintCredential_ProviderNotFound(t *testing.T) {
	ctx := context.Background()

	r, err := registry.NewRegistry(ctx, nil)
	require.NoError(t, err)

	_, err = r.MintCredential(ctx, types.CredentialMintRequest{
		Provider: types.ProviderType("nonexistent"),
	})

	assert.ErrorIs(t, err, registry.ErrProviderNotFound)
}

func TestRegistry_UpsertProvider_ProviderTypeRequired(t *testing.T) {
	ctx := context.Background()

	r, err := registry.NewRegistry(ctx, nil)
	require.NoError(t, err)

	err = r.UpsertProvider(ctx, spec.ProviderSpec{}, nil)
	assert.ErrorIs(t, err, registry.ErrProviderTypeRequired)
}

func TestRegistry_ResolveOperation_ProviderNotFound(t *testing.T) {
	ctx := context.Background()

	r, err := registry.NewRegistry(ctx, nil)
	require.NoError(t, err)

	_, err = r.ResolveOperation(types.ProviderType("nonexistent"), types.OperationHealthDefault, "")
	assert.ErrorIs(t, err, registry.ErrOperationNotRegistered)
}

func TestRegistry_ResolveOperation_CriteriaRequired(t *testing.T) {
	ctx := context.Background()

	r, err := registry.NewRegistry(ctx, nil)
	require.NoError(t, err)

	_, err = r.ResolveOperation(types.ProviderType("slack"), "", "")
	assert.ErrorIs(t, err, registry.ErrOperationCriteriaRequired)
}

func TestRegistry_ResolveOperation_ProviderUnknown(t *testing.T) {
	ctx := context.Background()

	r, err := registry.NewRegistry(ctx, nil)
	require.NoError(t, err)

	_, err = r.ResolveOperation(types.ProviderUnknown, types.OperationHealthDefault, "")
	assert.ErrorIs(t, err, registry.ErrProviderTypeRequired)
}
