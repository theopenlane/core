package operations

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/pkg/gala"
)

// RegisterListeners attaches one Gala listener per registered operation topic.
func RegisterListeners(runtime *gala.Gala, reg *registry.Registry, handle func(context.Context, Envelope) error) error {
	if runtime == nil {
		return ErrGalaRequired
	}

	for _, operation := range reg.Listeners() {
		if _, err := gala.RegisterListeners(runtime.Registry(), gala.Definition[Envelope]{
			Topic: gala.Topic[Envelope]{
				Name: operation.Topic,
			},
			Name: operation.Name,
			Handle: func(ctx gala.HandlerContext, envelope Envelope) error {
				return handle(ctx.Context, envelope)
			},
		}); err != nil {
			return err
		}
	}

	return nil
}
