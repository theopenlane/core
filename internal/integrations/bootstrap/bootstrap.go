package bootstrap

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/config"
	"github.com/theopenlane/core/internal/integrations/providers/catalog"
	"github.com/theopenlane/core/internal/integrations/registry"
)

// LoadRegistry loads provider specs using the supplied loader and builder catalog.
func LoadRegistry(ctx context.Context, loader *config.FSLoader) (*registry.Registry, error) {
	specs, err := loader.Load(ctx)
	if err != nil {
		return nil, err
	}

	builders := catalog.Builders()

	return registry.New(ctx, specs, builders)
}

// LoadDefaultRegistry loads provider specs from the embedded filesystem.
func LoadDefaultRegistry(ctx context.Context) (*registry.Registry, error) {
	loader := config.NewFSLoader(config.ProvidersFS, "providers")

	return LoadRegistry(ctx, loader)
}
