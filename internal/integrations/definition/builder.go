package definition

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder builds one manifest-backed definition
type Builder interface {
	// Build returns one fully-assembled definition
	Build(ctx context.Context) (types.Definition, error)
}

// BuilderFunc adapts a function into a definition builder
type BuilderFunc func(ctx context.Context) (types.Definition, error)

// Build returns one fully-assembled definition
func (f BuilderFunc) Build(ctx context.Context) (types.Definition, error) {
	if f == nil {
		return types.Definition{}, ErrBuilderNil
	}

	return f(ctx)
}

// BuildAll builds every supplied definition in order
func BuildAll(ctx context.Context, builders ...Builder) ([]types.Definition, error) {
	out := make([]types.Definition, 0, len(builders))

	for _, builder := range builders {
		if builder == nil {
			return nil, ErrBuilderNil
		}

		definition, err := builder.Build(ctx)
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
