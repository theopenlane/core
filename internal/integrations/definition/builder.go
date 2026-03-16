package definition

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder builds one manifest-backed definition
type Builder func(ctx context.Context) (types.Definition, error)

// BuildAll builds every supplied definition in order
func BuildAll(ctx context.Context, builders ...Builder) ([]types.Definition, error) {
	out := make([]types.Definition, 0, len(builders))

	for _, builder := range builders {
		if builder == nil {
			return nil, ErrBuilderNil
		}

		definition, err := builder(ctx)
		if err != nil {
			return nil, err
		}

		out = append(out, definition)
	}

	return out, nil
}

// RegisterAll builds and registers every supplied definition in order
func RegisterAll(ctx context.Context, reg *registry.Registry, builders ...Builder) error {
	if reg == nil {
		return ErrRegistryRequired
	}

	definitions, err := BuildAll(ctx, builders...)
	if err != nil {
		return err
	}

	for _, definition := range definitions {
		if err := reg.Register(definition); err != nil {
			return err
		}
	}

	return nil
}
